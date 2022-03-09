package configuration

import (
	"os"
	"path/filepath"
	"returntypes-langserver/common/dataformat/jsonschema"
)

type TypeClassConfigurations []TypeClass

// Contains all configured type classes
type TypeClassConfigurationFile struct {
	// The configured type classes
	Classes []TypeClass `json:"classes"`
}

type TypeClass struct {
	// The name of the type class
	Label string `json:"label"`
	// If true, then this type class is used as a type for all array types (may be only used for max. of one type class)
	IsArrayType bool `json:"isArrayType"`
	// If true, then this type class is used as a type for all methods which are chain methods (may be only used for max. of one type class)
	IsChainMethodType bool `json:"isChainMethodType"`
	// A list of canonical names of classes/types which belong to this type class including the ones extending or implementing them
	Elements []string `json:"elements"`
	// The color used for this type class for visualization
	Color string `json:"color"`
}

func (c TypeClassConfigurations) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if value == nil {
		value = filepath.Join(GoProjectDir(), "resources", "data", "typeClasses.json")
	}
	if filePath, ok := value.(string); ok {
		// Load configuration from different JSON file
		err = c.fromFilePath(filePath)
		value = c
	}
	return value, err
}

func (c *TypeClassConfigurations) fromFilePath(filePath string) error {
	if filePath == "" {
		return nil
	}
	contents, err := os.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	return c.fromJson(contents)
}

func (c *TypeClassConfigurations) fromJson(contents []byte) error {
	var config TypeClassConfigurationFile
	if err := jsonschema.UnmarshalJSONStrict(contents, &config, TypeClassConfigurationFileSchema); err != nil {
		return err
	}
	*c = config.Classes
	return nil
}
