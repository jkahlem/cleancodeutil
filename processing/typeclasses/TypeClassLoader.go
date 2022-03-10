package typeclasses

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

var loadedTypeClasses *TypeClassConfiguration

// Returns the loaded type class configuration
func GetTypeClasses() TypeClassConfiguration {
	if loadedTypeClasses == nil {
		return TypeClassConfiguration{}
	}
	return *loadedTypeClasses
}

// Loads the type class configuration
func LoadTypeClasses() errors.Error {
	if typeClasses, err := loadTypeClassesFromJsonFile(configuration.DefaultTypeClasses()); err != nil {
		return err
	} else if err := validateTypeClasses(typeClasses); err != nil {
		return err
	} else {
		loadedTypeClasses = typeClasses
		return nil
	}
}

// Checks if the type class configuration is valid
func validateTypeClasses(config *TypeClassConfiguration) errors.Error {
	if config == nil {
		return errors.New(TypeClassMapperErrorTitle, "Invalid type class configuration")
	}
	loadedTypes := make(utils.StringSet)
	for i, typeClass := range config.Classes {
		if typeClass.IsArrayType {
			if config.ArrayType != nil {
				return errors.New(TypeClassMapperErrorTitle, "An array type configuration is only for a maximum of one type class allowed.")
			}
			config.ArrayType = &config.Classes[i]
		}
		if typeClass.IsChainMethodType {
			if config.ChainMethodType != nil {
				return errors.New(TypeClassMapperErrorTitle, "A chain method type configuration is only for a maximum of one type class allowed.")
			}
			config.ChainMethodType = &config.Classes[i]
		}
		for _, typeName := range typeClass.Elements {
			if typeName == DefaultType {
				config.DefaultType = &config.Classes[i]
			}
			if loadedTypes.Has(typeName) {
				return errors.New(TypeClassMapperErrorTitle, fmt.Sprintf("The type %s is contained in different type classes. (Types must be unique in the type classes)", typeName))
			}
			loadedTypes.Put(typeName)
		}
	}
	if config.DefaultType == nil {
		return errors.New(TypeClassMapperErrorTitle, fmt.Sprintf("At least one type class needs to include the default type %s.", DefaultType))
	}
	return nil
}

// Loads the type class configuration from a json file
func loadTypeClassesFromJsonFile(path string) (*TypeClassConfiguration, errors.Error) {
	typeClasses := TypeClassConfiguration{}
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, TypeClassMapperErrorTitle, "Could not load type class configuration file")
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, TypeClassMapperErrorTitle, "Could not load type class configuration file")
	}
	if err := json.Unmarshal(content, &typeClasses); err != nil {
		return nil, errors.Wrap(err, TypeClassMapperErrorTitle, "Could not parse type class configuration file")
	}
	return &typeClasses, nil
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
				return config, errors.New(TypeClassMapperErrorTitle, fmt.Sprintf("The type %s is contained in different type classes. (Types must be unique in the type classes)", typeName))
			}
			loadedTypes.Put(typeName)
		}
	}
	if config.DefaultType == nil {
		return config, errors.New(TypeClassMapperErrorTitle, fmt.Sprintf("At least one type class needs to include the default type %s.", DefaultType))
	}
	return config, nil
}
