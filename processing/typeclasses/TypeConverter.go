package typeclasses

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
)

const TypeClassMapperErrorTitle = "Type Class Mapper Error"
const DefaultType = "java.lang.Object"
const UnknownType = "<Unknown>" // shown if the mapping has failed.

// The mapper is for mapping a type name to it's type class
type Mapper interface {
	// The package tree the mapper will use for type class mapping
	SetPackageTree(*packagetree.Tree)
	// Returns the type class of a method's return type
	MapReturnTypeToTypeClass(typeName string, methodLabels []string) (string, errors.Error)
	// Maps a parameter type to a type class
	MapParameterTypeToTypeClass(typeName string, methodLabels []string) (string, errors.Error)
}

// Maps a type to it's type class
type typeClassMapper struct {
	tree     *packagetree.Tree
	mappings map[string]string
	config   TypeClassConfiguration
}

func GetTypeClasses() TypeClassConfiguration {
	return TypeClassConfiguration{}
}

func New(tree *packagetree.Tree, typeclasses configuration.TypeClassConfigurations) (Mapper, errors.Error) {
	if config, err := buildTypeClassConfiguration(typeclasses); err != nil {
		return nil, err
	} else {
		mapper := typeClassMapper{tree: tree}
		mapper.loadTypeClasses(config)
		return &mapper, nil
	}
}

func (m *typeClassMapper) MapParameterTypeToTypeClass(typeName string, methodLabels []string) (string, errors.Error) {
	if m.tree == nil {
		return UnknownType, errors.New(TypeClassMapperErrorTitle, "No package tree set")
	} else if m.config.ArrayType != nil && m.findMethodLabel(methodLabels, java.ArrayType) {
		return m.config.ArrayType.Label, nil
	}
	return m.mapTypeToTypeClass(typeName), nil
}

func (m *typeClassMapper) MapReturnTypeToTypeClass(typeName string, methodLabels []string) (string, errors.Error) {
	if m.tree == nil {
		return UnknownType, errors.New(TypeClassMapperErrorTitle, "No package tree set")
	} else if m.config.ChainMethodType != nil && m.findMethodLabel(methodLabels, java.ChainMethod) {
		return m.config.ChainMethodType.Label, nil
	} else if m.config.ArrayType != nil && m.findMethodLabel(methodLabels, java.ArrayType) {
		return m.config.ArrayType.Label, nil
	}

	return m.mapTypeToTypeClass(typeName), nil
}

func (m *typeClassMapper) findMethodLabel(allLabels []string, labelToFind java.MethodLabel) bool {
	for _, methodLabel := range allLabels {
		if methodLabel == string(labelToFind) {
			return true
		}
	}
	return false
}

func (m *typeClassMapper) SetPackageTree(tree *packagetree.Tree) {
	m.tree = tree
}

// Loads the type classes into a map with type names -> type class
func (m *typeClassMapper) loadTypeClasses(typeClasses TypeClassConfiguration) {
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
