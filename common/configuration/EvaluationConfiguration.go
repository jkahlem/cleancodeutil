package configuration

import (
	"io/ioutil"
	"returntypes-langserver/common/dataformat/jsonschema"
)

type EvaluationConfiguration struct {
	// Subsets of the evaluation set for which scores should be also calculated (e.g. filter out setter/getter for evaluation and so on)
	Subsets []EvaluationSet `json:"subsets"`
}

func (c EvaluationConfiguration) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if filePath, ok := value.(string); ok {
		// Load configuration from different JSON file
		err = c.fromFilePath(filePath)
		value = c
	}
	return value, err
}

func (c *EvaluationConfiguration) fromFilePath(filePath string) error {
	if filePath == "" {
		return nil
	}
	contents, err := ioutil.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	var fileConfig EvaluationConfiguration
	if err := jsonschema.UnmarshalJSONStrict(contents, &fileConfig, EvaluationConfigurationFileSchema); err != nil {
		return err
	}
	*c = fileConfig
	return nil
}

type EvaluationSet struct {
	Subsets []EvaluationSet `json:"subsets"`
	// Defines, how the rating per row should be done, like equality checks or different tools etc.
	RatingTypes []string `json:"ratingTypes"`
	Filter      Filter   `json:"filter"`
}
