package typeclasses

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

// TODO: This class should either only work with string -> string convertions, or should be in a different package?

const TypeClassMapperErrorTitle = "Type Class Mapper Error"
const DefaultType = "java.lang.Object"
const UnknownType = "<Unknown>" // shown if the mapping has failed.

// The mapper is for mapping a type name to it's type class
type Mapper interface {
	// The package tree the mapper will use for type class mapping
	SetPackageTree(*packagetree.Tree)
	// Returns the type class of a method's return type
	MapReturnTypeToTypeClass(csv.Method) (string, errors.Error)
	// Maps all types of the methods to a type class
	MapMethodsTypesToTypeClass([]csv.Method) ([]csv.Method, errors.Error)
}

// Maps a type to it's type class
type typeClassMapper struct {
	tree     *packagetree.Tree
	mappings map[string]string
	config   TypeClassConfiguration
}

// Creates and prepares a new type class mapper
func New(tree *packagetree.Tree) Mapper {
	mapper := typeClassMapper{tree: tree}
	mapper.setup()
	return &mapper
}

func (m *typeClassMapper) SetPackageTree(tree *packagetree.Tree) {
	m.tree = tree
}

func (m *typeClassMapper) setup() {
	m.loadTypeClasses()
}

// Loads the default type classes
func (m *typeClassMapper) loadTypeClasses() {
	m.mapTypeClasses(GetTypeClasses())
}

// Loads the type classes into a map with type names -> type class
func (m *typeClassMapper) mapTypeClasses(typeClasses TypeClassConfiguration) {
	if m.mappings == nil {
		m.mappings = make(map[string]string)
	}
	for _, typeClass := range typeClasses.Classes {
		for _, typeName := range typeClass.Elements {
			m.mappings[typeName] = typeClass.Label
		}
	}
	m.config = typeClasses
}

// Maps the return type of multiple methods to it's type class
//
// Deprecated: Use MapMethodsTypesToTypeClass instead ...
func (m *typeClassMapper) MapReturnTypesToTypeClass(methods []csv.Method) ([]csv.Method, errors.Error) {
	if m.tree == nil {
		return nil, errors.New(TypeClassMapperErrorTitle, "No package tree set")
	}
	result := make([]csv.Method, len(methods))
	for i, method := range methods {
		if returnType, err := m.MapReturnTypeToTypeClass(method); err != nil {
			return nil, err
		} else {
			result[i] = method
			result[i].ReturnType = returnType
		}
	}
	return result, nil
}

// Maps the return type of multiple methods to it's type class
func (m *typeClassMapper) MapMethodsTypesToTypeClass(methods []csv.Method) ([]csv.Method, errors.Error) {
	if m.tree == nil {
		return nil, errors.New(TypeClassMapperErrorTitle, "No package tree set")
	}
	result := make([]csv.Method, len(methods))
	for i, method := range methods {
		returnType, err := m.MapReturnTypeToTypeClass(method)
		if err != nil {
			return nil, err
		}
		result[i] = method
		result[i].ReturnType = returnType
		result[i].Parameters = method.Parameters // m.mapParameterTypesToTypeClass(method.Parameters)
	}
	return result, nil
}

// maps the parameters to have a type class instead of the type name ...
func (m *typeClassMapper) mapParameterTypesToTypeClass(parameters []string) []string {
	if csv.IsEmptyList(parameters) {
		return nil
	}
	results := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		// splitted has for each element the pattern "<type> <name>"
		splitted := strings.Split(parameter, " ")
		splitted[0] = m.mapTypeToTypeClass(splitted[0])
		results = append(results, strings.Join(splitted, " "))
	}
	return results
}

// Returns the name of the type class for the method's return type
func (m *typeClassMapper) MapReturnTypeToTypeClass(method csv.Method) (string, errors.Error) {
	if m.tree == nil {
		return UnknownType, errors.New(TypeClassMapperErrorTitle, "No package tree set")
	}
	if m.config.ChainMethodType != nil && method.HasLabel(string(java.ChainMethod)) {
		return m.config.ChainMethodType.Label, nil
	}
	if m.config.ArrayType != nil && method.HasLabel(string(java.ArrayType)) {
		return m.config.ArrayType.Label, nil
	}

	return m.mapTypeToTypeClass(method.ReturnType), nil
}

// Returns the type class of the given type
func (m *typeClassMapper) mapTypeToTypeClass(typeName string) string {
	if name, found := m.mapByTypeClassElement(typeName); found {
		return name
	}
	if name, found := m.mapByParentClass(typeName); found {
		return name
	}
	if typeName == DefaultType {
		// if the default type comes until here, then there is a configuration miss, so return unknown type
		return UnknownType
	}
	// Otherwise use the default type as fallback. This is for example the case if a type is not in the class hierarchy because it
	// does not extend any type (except for java.lang.Object) or because they are inside external dependencies.
	fallbackType, _ := m.mapByTypeClassElement(DefaultType)
	return fallbackType
}

// Maps the type directly to a type class if found (so type name is directly in the type class)
func (m *typeClassMapper) mapByTypeClassElement(typeName string) (string, bool) {
	if typeClass, ok := m.mappings[typeName]; ok {
		return typeClass, true
	}
	return UnknownType, false
}

// Maps the type to the type class the parent class belongs to
func (m *typeClassMapper) mapByParentClass(typeName string) (string, bool) {
	// prevent endless loops when searching for default type
	if typeName == DefaultType {
		return DefaultType, true
	}

	// search for java class node in package tree
	selector := m.tree.Select(typeName)
	if node := selector.Get(); node != nil {
		if class, ok := node.(*java.Class); ok {
			// if the class has no parent classes, use the default type (java.lang.Object)
			if len(class.ExtendsImplements) == 0 {
				return m.mapTypeToTypeClass(DefaultType), true
			}

			resolvedType, _ := java.Resolve(&class.ExtendsImplements[0], m.tree)
			return m.mapTypeToTypeClass(resolvedType), true
		}
	}
	return UnknownType, false
}
