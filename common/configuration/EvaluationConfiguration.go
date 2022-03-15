package configuration

import (
	"fmt"
	"io/ioutil"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
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
	Metrics []MetricConfiguration `json:"metrics"`
	Filter  Filter                `json:"filter"`
}

const (
	RougeL = "rouge-l"
	RougeS = "rouge-s"
	RougeN = "rouge-n"
	Bleu   = "bleu"
)

type MetricConfiguration map[string]interface{}

func (c MetricConfiguration) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if metricType, ok := value.(string); ok {
		switch metricType {
		case RougeL:
			return RougeLConfiguration{
				Type: RougeL,
			}, nil
		case RougeS:
			return RougeSConfiguration{
				Type: RougeS,
			}, nil
		case "rouge-2":
			return RougeNConfiguration{
				Type: RougeN,
				N:    2,
			}, nil
		case Bleu:
			return BleuConfiguration{
				Type: Bleu,
			}, nil
		default:
			return value, fmt.Errorf("Unsupported metric type preset: %s", metricType)
		}
	}
	return value, err
}

func (c MetricConfiguration) AsRougeL() (RougeLConfiguration, errors.Error) {
	var config RougeLConfiguration
	err := c.as(RougeL, &config)
	return config, err
}

func (c MetricConfiguration) AsRougeS() (RougeSConfiguration, errors.Error) {
	var config RougeSConfiguration
	err := c.as(RougeS, &config)
	return config, err
}

func (c MetricConfiguration) AsRougeN() (RougeNConfiguration, errors.Error) {
	var config RougeNConfiguration
	err := c.as(RougeN, &config)
	return config, err
}

func (c MetricConfiguration) AsBleu() (BleuConfiguration, errors.Error) {
	var config BleuConfiguration
	err := c.as(Bleu, &config)
	return config, err
}

func (c MetricConfiguration) as(expectedType string, destination interface{}) errors.Error {
	if val, ok := c["type"]; !ok || val != expectedType {
		return errors.New("Type Error", fmt.Sprintf("Cannot interpret metric type '%s' as %s", val, expectedType))
	}
	return utils.DecodeMapToStructStrict(c, &destination)
}

func (c MetricConfiguration) Type() string {
	if value, ok := c["type"]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

type RougeLConfiguration struct {
	Type     string    `json:"type"`
	Measures []Measure `json:"measure"`
}

type RougeSConfiguration struct {
	Type     string    `json:"type"`
	SkipN    int       `json:"skipN"`
	Measures []Measure `json:"measure"`
}

type RougeNConfiguration struct {
	Type     string    `json:"type"`
	N        int       `json:"n"`
	Measures []Measure `json:"measure"`
}

type BleuConfiguration struct {
	Type    string    `json:"type"`
	Weights []float64 `json:"weights"`
}

const FScore = "fscore"

type Measure map[string]interface{}

func (c Measure) DecodeValue(value interface{}) (interface{}, error) {
	var err error
	if metricType, ok := value.(string); ok {
		switch metricType {
		case "f1score":
			return FScoreConfiguration{
				Type: FScore,
				Beta: 1,
			}, nil
		default:
			return value, fmt.Errorf("Unsupported measure type preset: %s", metricType)
		}
	}
	return value, err
}

type FScoreConfiguration struct {
	Type string  `json:"type"`
	Beta float64 `json:"beta"`
}

func (c Measure) AsFScore() (FScoreConfiguration, errors.Error) {
	var config FScoreConfiguration
	err := c.as(FScore, &config)
	return config, err
}

func (c Measure) as(expectedType string, destination interface{}) errors.Error {
	if val, ok := c["type"]; !ok || val != expectedType {
		return errors.New("Type Error", fmt.Sprintf("Cannot interpret metric type '%s' as %s", val, expectedType))
	}
	return utils.DecodeMapToStructStrict(c, &destination)
}

func (c Measure) Type() string {
	if value, ok := c["type"]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}
