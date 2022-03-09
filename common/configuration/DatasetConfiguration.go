package configuration

import (
	"os"
	"returntypes-langserver/common/dataformat/jsonschema"
)

type DatasetConfiguration []Dataset

type DatasetFileConfiguration struct {
	Datasets []Dataset `json:"datasets"`
}

type Dataset struct {
	Name           string         `json:"name"`
	Filter         Filter         `json:"filter"`
	IsGroupOnly    bool           `json:"isGroupOnly"`
	Description    string         `json:"description"`
	SpecialOptions SpecialOptions `json:"specialOptions"`
	ModelOptions   ModelOptions   `json:"modelOptions"`
	Subsets        []Dataset      `json:"subsets"`
}

type SpecialOptions struct {
	Convert2And4ToWords bool `json:"convert2And4ToWords"`
	MinMethodNameLength int  `json:"minMethodNameLength"`
	FilterDuplicates    bool `json:"filterDuplicates"`
	// TODO: Actually load and validate typeclasses for this one?
	TypeClasses string `json:"typeClasses"`
	// The size of the splitted datasets as a proportion
	DatasetSize DatasetProportion `json:"datasetSize"`
}

type ModelOptions struct {
}

type ModelType string

const (
	MethodGenerator      ModelType = "MethodGenerator"
	ReturnTypesValidator ModelType = "ReturnTypesValidator"
)

func (c DatasetConfiguration) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if filePath, ok := value.(string); ok {
		// Load configuration from different JSON file
		err = c.fromFilePath(filePath)
		value = c
	}
	return value, err
}

func (c *DatasetConfiguration) fromFilePath(filePath string) error {
	contents, err := os.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	return c.fromJson(contents)
}

func (c *DatasetConfiguration) fromJson(contents []byte) error {
	var config DatasetFileConfiguration
	if err := jsonschema.UnmarshalJSONStrict(contents, &config, ExcelSetConfigurationFileSchema); err != nil {
		return err
	}
	*c = config.Datasets
	return nil
}
