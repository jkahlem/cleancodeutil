package java

import (
	"strings"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/packagetree"
	"returntypes-langserver/common/utils/counter"
)

type ResolutionState int

const (
	Unresolved     ResolutionState = 0
	ReferenceFound ResolutionState = 3
	Resolved       ResolutionState = 4
	InProgress     ResolutionState = 5
)

const DependencyImportCounter = "resolverDependencyImportCounter"
const UnresolvedTypeCounter = "resolverUnresolvedTypeCounter"

// Resolves type names in java to their canonical names.
type Resolver struct {
	targetType *Type
	start      JavaElement
	path       string
	err        errors.Error
	tree       *packagetree.Tree
	// If true, the resolver will not ensure that types will be updated if the tree changes.
	// This will speed up the resolution process where tree updates do not exist (e.g. when creating the dataset)
	NoUpdatesOnTreeChange bool
}

// Sets the target type to resolve.
func (resolver *Resolver) SetTarget(javaType *Type) {
	resolver.targetType = javaType
}

// Sets the point from which the resolution process starts (the resolver will go upwards).
func (resolver *Resolver) SetStartPoint(element JavaElement) {
	resolver.start = element
}

// The resolution state of the target type.
func (resolver *Resolver) State() ResolutionState {
	if resolver.targetType == nil {
		return Unresolved
	}
	return resolver.targetType.TypeResolutionState
}

// The (qualified) name of the target type.
func (resolver *Resolver) TypeName() string {
	if resolver.targetType == nil {
		return ""
	}
	return resolver.targetType.TypeName
}

// Returns a list of each parts of the (qualified) type name.
func (resolver *Resolver) TypeNameSplitted() []string {
	return strings.Split(resolver.TypeName(), ".")
}

// Returns the canonical name after resolution.
func (resolver *Resolver) ResolvedTypeName() string {
	if resolver.targetType == nil {
		return ""
	} else if resolver.targetType.resolutionTypeName == "" {
		return resolver.TypeName()
	}
	return resolver.targetType.resolutionTypeName
}

// Returns true if the type was resolved.
func (resolver *Resolver) IsResolved() bool {
	return resolver.State() == Resolved
}

// Starts the resolution process.
func (resolver *Resolver) Resolve() {
	if resolver.targetType == nil ||
		resolver.targetType.TypeResolutionState == Resolved ||
		resolver.targetType.TypeResolutionState == InProgress {
		return
	}

	// Set in progress to prevent endless loops
	resolver.targetType.TypeResolutionState = InProgress

	current := resolver.start
	if current == nil {
		current = resolver.targetType.Parent()
	}

	resolver.resolvePrimitive()

	// Check if the typename is already a fully qualified path
	resolver.resolveByPath(resolver.TypeName())
	for current != nil && !resolver.IsResolved() {
		resolver.resolveElement(current)
		current = current.Parent()
	}
}

// Resolve the type by the given element.
func (resolver *Resolver) resolveElement(element JavaElement) {
	switch t := element.(type) {
	case *Method:
		resolver.resolveMethod(t)
	case *TypeParameter:
		resolver.resolveTypeParameter(t)
	case *Class:
		resolver.resolveClass(t)
	case *CodeFile:
		resolver.resolveFile(t)
	case *Import:
		resolver.resolveImport(t)
	}
}

// Search for an element at the defined path and resolve to it if it exist.
func (resolver *Resolver) resolveByPath(path string) {
	if resolver.IsResolved() {
		return
	}

	selector := resolver.tree.Select(path)
	if node := selector.Get(); node != nil {
		if element, ok := node.(JavaElement); ok {
			resolver.resolveTo(element)
			return
		}
	}
}

// Try to resolve to a primitive type.
func (resolver *Resolver) resolvePrimitive() {
	if resolver.targetType == nil || resolver.IsResolved() {
		return
	}

	if len(resolver.TypeNameSplitted()) == 1 {
		switch resolver.TypeName() {
		case "byte", "short", "int", "long", "float", "double", "char", "boolean", "void":
			resolver.targetType.TypeResolutionState = Resolved
			resolver.targetType.resolutionTypeName = resolver.TypeName()
		}
	}
}

// Try to resolve to the method's type parameters.
func (resolver *Resolver) resolveMethod(method *Method) {
	if resolver.IsResolved() {
		return
	}

	for i := range method.TypeParameters {
		resolver.resolveTypeParameter(&method.TypeParameters[i])
		if resolver.IsResolved() {
			return
		}
	}
}

// Try to resolve to a type representing the type parameter. By default, this will be java.lang.Object as a type parameter can be any class.
// If a type bound is set, the type will resolve to it as classes used for the type parameter must extend or implement the type of the type bound.
//
// The resolution to a type boudn will only apply to the first type bound as currently resolutions to multiple types are not supported.
func (resolver *Resolver) resolveTypeParameter(typeParameter *TypeParameter) {
	if resolver.IsResolved() || len(resolver.TypeNameSplitted()) != 1 ||
		resolver.TypeName() != typeParameter.TypeParameterName {
		return
	}

	if len(typeParameter.TypeBounds) > 0 {
		boundResolver := Resolver{tree: resolver.tree}
		boundResolver.SetTarget(&typeParameter.TypeBounds[0])
		boundResolver.SetStartPoint(typeParameter.Parent())
		boundResolver.Resolve()

		if typeParameter.TypeBounds[0].TypeResolutionState == Resolved {
			resolver.targetType.resolutionTypeName = typeParameter.TypeBounds[0].TypeName
			resolver.targetType.TypeResolutionState = typeParameter.TypeBounds[0].TypeResolutionState
		}
	} else {
		resolver.targetType.resolutionTypeName = "java.lang.Object"
		resolver.targetType.TypeResolutionState = Resolved
	}
}

// Try to resolve the type to type parameters and sub classes of the class and of the classes it extends.
func (resolver *Resolver) resolveClass(class *Class) {
	if resolver.targetType == nil || resolver.tree == nil || resolver.IsResolved() {
		return
	}

	resolver.resolveClassUsingPathInPackageTree(class)
	resolver.resolveClassByTypeParameters(class)
	resolver.resolveClassByLookingInExtendedClasses(class)
}

// Search for the class in the package tree by appending the type name to the parent's path.
func (resolver *Resolver) resolveClassUsingPathInPackageTree(class *Class) {
	if resolver.IsResolved() {
		return
	}

	targetPath := strings.Join([]string{class.Path(), resolver.TypeName()}, ".")
	if class.Parent() != nil && resolver.TypeNameSplitted()[0] == class.ClassName {
		// if the typename starts with this classes name, add the path to the parent
		targetPath = strings.Join([]string{class.Parent().Path(), resolver.TypeName()}, ".")
	}

	resolver.resolveByPath(targetPath)
}

// Try to resolve the type to the classes type parameters.
func (resolver *Resolver) resolveClassByTypeParameters(class *Class) {
	if resolver.IsResolved() {
		return
	}

	for i := range class.TypeParameters {
		resolver.resolveTypeParameter(&class.TypeParameters[i])
		if resolver.IsResolved() {
			return
		}
	}
}

// Search for extended classes and continue the resolution process in the context of the extended classes.
func (resolver *Resolver) resolveClassByLookingInExtendedClasses(class *Class) {
	if resolver.IsResolved() {
		return
	}

	for i := range class.ExtendsImplements {
		extendedClassResolver := Resolver{tree: resolver.tree}
		extendedClassResolver.SetStartPoint(class.Parent())
		extendedClassResolver.SetTarget(&class.ExtendsImplements[i])
		extendedClassResolver.Resolve()
		if extendedClassResolver.State() != Resolved {
			continue
		}

		resolver.resolveByPath(strings.Join([]string{extendedClassResolver.ResolvedTypeName(), resolver.TypeName()}, "."))
	}
}

// Try to resolve the type on file level:
// - to imported types
// - to types in the same package
// - to types of the java.lang package
func (resolver *Resolver) resolveFile(codeFile *CodeFile) {
	if resolver.targetType == nil || resolver.tree == nil || resolver.IsResolved() {
		return
	}

	resolver.resolveByImportsOfFile(codeFile, false)
	resolver.resolveByPackage(codeFile.PackageName)
	resolver.resolveByImportsOfFile(codeFile, true)
	resolver.resolveByPackage("java.lang")
}

// Try to resolve the type to imported types.
// If byWildcards is true, the type will be resolved to types imported by wildcard imports like:
//   import java.util.*;
// Otherwise the type will will be resolved to types imported by single imports like:
//   import java.util.List;
func (resolver *Resolver) resolveByImportsOfFile(codeFile *CodeFile, byWildcards bool) {
	if resolver.IsResolved() {
		return
	}

	for i := range codeFile.Imports {
		if codeFile.Imports[i].IsWildcard == byWildcards {
			resolver.resolveImport(&codeFile.Imports[i])
			if resolver.IsResolved() {
				return
			}
		}
	}
}

// Try to resolve the type to an imported type.
func (resolver *Resolver) resolveImport(_import *Import) {
	if resolver.targetType == nil || resolver.tree == nil || resolver.IsResolved() {
		return
	}

	if _import.IsWildcard {
		fullPath := strings.Join([]string{_import.ImportPath, resolver.TypeName()}, ".")
		resolver.resolveByPath(fullPath)
	} else if resolver.importPathEndsWithTypeName(_import) {
		// merge import path and type name (they have the same last element)
		pathSplitted := resolver.TypeNameSplitted()
		pathSplitted[0] = _import.ImportPath
		newPath := strings.Join(pathSplitted, ".")
		resolver.resolveByPath(newPath)

		if !resolver.IsResolved() {
			// there is no node representing the path (because it is from an external source not in the package tree etc.)
			resolver.targetType.resolutionTypeName = strings.Join(pathSplitted, ".")
			resolver.targetType.TypeResolutionState = Resolved
			resolver.subscribeTypeToRootNode()
			counter.For(DependencyImportCounter).CountUp()
		}
	}
}

// Returns true if the import path ends with the type name.
func (resolver *Resolver) importPathEndsWithTypeName(_import *Import) bool {
	splittedImport := strings.Split(_import.ImportPath, ".")
	return splittedImport[len(splittedImport)-1] == resolver.TypeNameSplitted()[0]
}

// Try to resolve a type to a type in the package.
func (resolver *Resolver) resolveByPackage(packageName string) {
	if resolver.IsResolved() {
		return
	}

	resolver.resolveByPath(strings.Join([]string{packageName, resolver.TypeName()}, "."))
}

// Shorthand for resolving java elements to their path.
func (resolver *Resolver) resolveTo(element JavaElement) {
	resolver.targetType.resolutionTypeName = element.Path()
	resolver.targetType.TypeResolutionState = Resolved
	resolver.subscribeTypeToElementCodeFile(element)
}

// Subscribes the type to the element's code file.
func (resolver *Resolver) subscribeTypeToElementCodeFile(element JavaElement) {
	if resolver.NoUpdatesOnTreeChange || resolver.targetType == nil {
		return
	}

	if codeFile := FindCodeFile(element); codeFile != nil {
		codeFile.Subscribe(resolver.targetType)
	}
}

// Subscribes the type to the tree's root node.
func (resolver *Resolver) subscribeTypeToRootNode() {
	if resolver.NoUpdatesOnTreeChange || resolver.targetType == nil || resolver.tree == nil || resolver.tree.Root == nil {
		return
	}
	resolver.tree.Root.AddSubscriber(resolver.targetType)
}

// Resolves a type to its canonical name.
func Resolve(javaType *Type, tree *packagetree.Tree) (resolvedTypeName string, isResolved bool) {
	resolver := Resolver{tree: tree}
	resolver.NoUpdatesOnTreeChange = true
	resolver.SetTarget(javaType)
	resolver.Resolve()
	if !resolver.IsResolved() {
		counter.For(UnresolvedTypeCounter).CountUp()
	}
	return resolver.ResolvedTypeName(), resolver.IsResolved()
}
