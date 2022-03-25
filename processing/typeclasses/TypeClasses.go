package typeclasses

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

// Contains all configured type classes
type TypeClassConfiguration struct {
	// The configured type classes
	Classes []configuration.TypeClass `json:"classes"`
	// A type class which is defined as an array type
	ArrayType *configuration.TypeClass
	// A type class which is defined as a chain method type
	ChainMethodType *configuration.TypeClass
	// A type class which contains java.lang.Object
	DefaultType *configuration.TypeClass
}

// Builds a type class configuration from type classes
func buildTypeClassConfiguration(typeClasses []configuration.TypeClass) (TypeClassConfiguration, errors.Error) {
	config := TypeClassConfiguration{
		Classes: typeClasses,
	}
	if typeClasses == nil {
		return config, errors.New(TypeClassMapperErrorTitle, "Invalid type class configuration")
	}
	loadedTypes := make(utils.StringSet)
	for i, typeClass := range config.Classes {
		if typeClass.IsArrayType {
			if config.ArrayType != nil {
				return config, errors.New(TypeClassMapperErrorTitle, "An array type configuration is only for a maximum of one type class allowed.")
			}
			config.ArrayType = &config.Classes[i]
		}
		if typeClass.IsChainMethodType {
			if config.ChainMethodType != nil {
				return config, errors.New(TypeClassMapperErrorTitle, "A chain method type configuration is only for a maximum of one type class allowed.")
			}
			config.ChainMethodType = &config.Classes[i]
		}
		for _, typeName := range typeClass.Elements {
			if typeName == DefaultType {
				config.DefaultType = &config.Classes[i]
			}
			if loadedTypes.Has(typeName) {
				return config, errors.New(TypeClassMapperErrorTitle, "The type %s is contained in different type classes. (Types must be unique in the type classes)", typeName)
			}
			loadedTypes.Put(typeName)
		}
	}
	if config.DefaultType == nil {
		return config, errors.New(TypeClassMapperErrorTitle, "At least one type class needs to include the default type %s.", DefaultType)
	}
	return config, nil
}
