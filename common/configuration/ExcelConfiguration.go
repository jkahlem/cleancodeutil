package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"
)

type ExcelSetConfiguration []ExcelSet

func (c *ExcelSetConfiguration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if filePath, ok := v.(string); ok {
		// Load configuration from different JSON file
		return c.fromFilePath(filePath)
	} else if slice, ok := v.([]interface{}); ok {
		return c.fromSlice(slice)
	} else {
		return fmt.Errorf("Unsupported project configuration value: %v", v)
	}
}

func (c ExcelSetConfiguration) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if filePath, ok := value.(string); ok {
		// Load configuration from different JSON file
		err = c.fromFilePath(filePath)
		value = c
	}
	return value, err
}

func (c *ExcelSetConfiguration) fromFilePath(filePath string) error {
	contents, err := os.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	return c.fromJson(contents)
}

func (c *ExcelSetConfiguration) fromJson(contents []byte) error {
	var config ExcelConfiguration
	if err := jsonschema.UnmarshalJSONStrict(contents, &config, ExcelSetConfigurationFileSchema); err != nil {
		return err
	}
	*c = config.ExcelSets
	return nil
}

func (c *ExcelSetConfiguration) fromSlice(slice []interface{}) error {
	*c = make(ExcelSetConfiguration, len(slice))
	for i, element := range slice {
		if err := (*c)[i].fromInterface(element); err != nil {
			return err
		}
	}
	return nil
}

type PatternType string

const (
	Wildcard PatternType = "wildcard"
	RegExp   PatternType = "regexp"
)

type ExcelConfiguration struct {
	ExcelSets []ExcelSet `json:"excelSets"`
}

type ExcelSet struct {
	Name               string     `json:"name"`
	Filter             Filter     `json:"filter"`
	NoOutput           bool       `json:"noOutput"`
	Subsets            []ExcelSet `json:"subsets"`
	ComplementFilename string     `json:"complementFilename"`
}

func (c *ExcelSet) fromInterface(itf interface{}) error {
	if jsonObj, ok := itf.(map[string]interface{}); ok {
		return utils.DecodeMapToStructStrict(jsonObj, c)
	} else {
		return fmt.Errorf("Unsupported excel set configuration value: %v", itf)
	}
}

func validateFilter(filter FilterConfiguration, datasetName string) errors.Error {
	for _, pattern := range filter.Method {
		if pattern.Type != RegExp && pattern.Pattern != strings.ToLower(pattern.Pattern) {
			return errors.New("Excel Error", fmt.Sprintf("Invalid method name pattern in dataset %s: Uppercase characters are not allowed.", datasetName))
		}
	}
	return nil
}

func validateFilters(filters FilterConfigurations, datasetName string) errors.Error {
	if filters != nil {
		for _, f := range filters {
			if err := validateFilter(f, datasetName); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDataset(dataset ExcelSet) errors.Error {
	if err := validateFilters(dataset.Filter.Includes, dataset.Name); err != nil {
		return err
	} else if err := validateFilters(dataset.Filter.Excludes, dataset.Name); err != nil {
		return err
	} else if err := validateDatasets(dataset.Subsets); err != nil {
		return err
	}
	return nil
}

func validateDatasets(datasets []ExcelSet) errors.Error {
	for _, subset := range datasets {
		if err := validateDataset(subset); err != nil {
			return err
		}
	}
	return nil
}
