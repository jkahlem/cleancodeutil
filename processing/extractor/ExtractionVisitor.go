package extractor

import (
	"fmt"
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/dataformat/csv"
	"strings"
)

// This visitor visists the elements of a Java Code File and writes each method/class found as a CSV record to the
// methods and classes properties. The type names are resolved to their canonical name (if resolution is possible)
type ExtractionVisitor struct {
	// List of CSV Records for java methods
	methods [][]string
	// List of CSV Records for java classes
	classes [][]string
	// Package tree to use for type resolution
	packageTree *packagetree.Tree
	// The currently visited java code file.
	currentFile *java.CodeFile
	// The currently visited class
	currentClass *java.Class
}

func (visitor *ExtractionVisitor) VisitCodeFile(codeFile *java.CodeFile) {
	visitor.currentFile = codeFile
	if codeFile.Classes != nil {
		for i := range codeFile.Classes {
			codeFile.Classes[i].Accept(visitor)
		}
	}
}

func (visitor *ExtractionVisitor) VisitClass(class *java.Class) {
	if class.Classes != nil {
		for i := range class.Classes {
			class.Classes[i].Accept(visitor)
		}
	}

	visitor.currentClass = class
	if class.Methods != nil {
		for i := range class.Methods {
			class.Methods[i].Accept(visitor)
		}
	}

	visitor.extractClass(class)
}

// extracts classes with their extended/implemented classes for the classHierarchyWriter using the fully qualified names
func (visitor *ExtractionVisitor) extractClass(class *java.Class) {
	if class.ExtendsImplements != nil {
		visitor.classes = append(visitor.classes, csv.Class{
			ClassName: visitor.currentFile.PackageName + "." + class.ClassName,
			Extends:   visitor.getExtendedClassesWithCanonicalName(class),
		}.ToRecord())
	}
}

// returns a list of the canonical name of each class extended/implemented by the given class
func (visitor *ExtractionVisitor) getExtendedClassesWithCanonicalName(class *java.Class) []string {
	extendedClasses := make([]string, len(class.ExtendsImplements))
	for index := range class.ExtendsImplements {
		extendedClassName := &class.ExtendsImplements[index]
		resolvedClassName, _ := visitor.resolve(extendedClassName)
		extendedClasses[index] = resolvedClassName
	}
	return extendedClasses
}

// resolves the type object to its canonical name. The bool return value is true if the type could be resolved.
func (visitor *ExtractionVisitor) resolve(typeName *java.Type) (string, bool) {
	return java.Resolve(typeName, visitor.packageTree)
}

// Writes methods with their return types into a csvfile
func (visitor *ExtractionVisitor) VisitMethod(method *java.Method) {
	resolvedReturnType, _ := visitor.resolve(&method.ReturnType)
	filePath := ""
	if visitor.currentFile != nil {
		filePath = visitor.currentFile.FilePath
	}
	visitor.methods = append(visitor.methods, csv.Method{
		MethodName: method.MethodName,
		ReturnType: resolvedReturnType,
		Parameters: visitor.mapParameters(method.Parameters),
		Labels:     java.GetMethodLabels(method),
		FilePath:   filePath,
		ClassName:  visitor.getQualifiedCurrentClassName(),
		Modifier:   method.Modifier,
		Exceptions: visitor.mapExceptions(method.Exceptions),
	}.ToRecord())
}

// Maps parameters in this format: "<type> <method>"
func (visitor *ExtractionVisitor) mapParameters(parameters []java.Parameter) []string {
	result := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		//resolvedParameterType, _ := visitor.resolve(&parameter.Type)
		unqualifiedTypeName := visitor.getUnqualifiedTypeName(parameter.Type.TypeName)
		csvStr := fmt.Sprintf("%s %s", unqualifiedTypeName, parameter.Name)
		result = append(result, csvStr)
	}
	return result
}

// Gets the qualified name of the class which is currently visited. The name includes all classes in which it is defined.
// Example:
//   class A {
//	   class B {}
//   }
// The return value for class B would be "A.B"
func (visitor *ExtractionVisitor) getQualifiedCurrentClassName() string {
	if visitor.currentClass == nil {
		return ""
	}
	return visitor.getQualifiedClassName(visitor.currentClass)
}

func (visitor *ExtractionVisitor) getQualifiedClassName(class *java.Class) string {
	if class == nil {
		return ""
	} else if upperClass, ok := class.Parent().(*java.Class); ok && upperClass != nil {
		return visitor.getQualifiedClassName(upperClass) + "." + class.ClassName
	} else {
		return class.ClassName
	}
}

// Maps exceptions into a string slice with their unqualified type names
func (visitor *ExtractionVisitor) mapExceptions(exceptions []java.Type) []string {
	result := make([]string, 0, len(exceptions))
	for _, exception := range exceptions {
		//resolvedParameterType, _ := visitor.resolve(&parameter.Type)
		result = append(result, visitor.getUnqualifiedTypeName(exception.TypeName))
	}
	return result
}

// Gets the unqualified name from an identifier (which might already be unqualified)
func (visitor *ExtractionVisitor) getUnqualifiedTypeName(identifier string) string {
	splitted := strings.Split(identifier, ".")
	return splitted[len(splitted)-1]
}

func (visitor *ExtractionVisitor) VisitImport(_import *java.Import) {
	// Do nothing.
}

func (visitor *ExtractionVisitor) VisitTypeParameter(typeParameter *java.TypeParameter) {
	// Do nothing.
}
