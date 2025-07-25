package configuration

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"
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
		if IsLangServMode() {
			// not relevant for language server
			return nil
		}
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
	Name         string                `json:"name"`
	Subsets      []EvaluationSet       `json:"subsets"`
	Metrics      []MetricConfiguration `json:"metrics"`
	Filter       Filter                `json:"filter"`
	TargetModels []string              `json:"targetModels"`
	Examples     []ExampleGroup        `json:"examples"`
}

type MethodExample struct {
	MethodName string `json:"methodName"`
	Static     bool   `json:"static"`
	ClassName  string `json:"className"`
	Label      string `json:"label"`
}

type ExampleGroup struct {
	Label    string          `json:"label"`
	Examples []MethodExample `json:"examples"`
}

var MethodExampleMatcher = regexp.MustCompile("^(.+:)?(static )?(([a-zA-Z][a-zA-Z0-9_]*\\.)*)([a-zA-Z][a-zA-Z0-9_]*)$")

func (g ExampleGroup) DecodeValue(value interface{}) (interface{}, error) {
	if pattern, ok := value.(string); ok {
		// The value is just a pattern
		match := MethodExampleMatcher.FindStringSubmatch(pattern)
		if len(match) != 6 {
			return nil, fmt.Errorf("could not parse method example pattern: '%s'", pattern)
		}
		if match[1] != "" {
			g.Label = match[1][:len(match[1])-1]
		}
		if example, err := g.decodeExample(value); err != nil {
			return value, err
		} else {
			g.Examples = []MethodExample{example}
		}
		return g, nil
	} else if jsonObj, ok := value.(map[string]interface{}); ok {
		if _, hasExample := jsonObj["examples"]; hasExample {
			// The value is already an example group
			return value, nil
		}
		// Otherwise, the value is a simple example definition
		if jsonObj["label"] != nil {
			if label, ok := jsonObj["label"].(string); ok {
				g.Label = label
			}
		}
		if example, err := g.decodeExample(value); err != nil {
			return value, err
		} else {
			g.Examples = []MethodExample{example}
		}
	}
	return value, nil
}

func (g ExampleGroup) decodeExample(value interface{}) (MethodExample, error) {
	decoded, err := (MethodExample{}).DecodeValue(value)
	if err != nil {
		return MethodExample{}, err
	} else if example, ok := decoded.(MethodExample); !ok {
		return MethodExample{}, fmt.Errorf("invalid example definition")
	} else {
		return example, nil
	}
}

func (e MethodExample) DecodeValue(value interface{}) (interface{}, error) {
	if pattern, ok := value.(string); ok {
		match := MethodExampleMatcher.FindStringSubmatch(pattern)
		if len(match) != 6 {
			return nil, fmt.Errorf("could not parse method example pattern: '%s'", pattern)
		}
		if match[2] != "" {
			e.Static = true
		}
		if match[3] != "" {
			e.ClassName = strings.TrimRight(match[3], ".")
		}
		e.MethodName = match[5]
		return e, nil
	}
	return value, nil
}

const (
	RougeL             = "rouge-l"
	RougeS             = "rouge-s"
	RougeN             = "rouge-n"
	Bleu               = "bleu"
	TokenCounter       = "tokenCounter"
	CompilabilityMatch = "compilability"
	ExactMatch         = "exactMatch"
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
		case TokenCounter:
			return TokenCounterConfiguration{
				Type: TokenCounter,
			}, nil
		case ExactMatch:
			return ExactMatchConfiguration{
				Type: ExactMatch,
			}, nil
		case CompilabilityMatch:
			return CompilabilityMatchConfiguration{
				Type: CompilabilityMatch,
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

func (c MetricConfiguration) AsTokenCounter() (TokenCounterConfiguration, errors.Error) {
	var config TokenCounterConfiguration
	err := c.as(TokenCounter, &config)
	return config, err
}

func (c MetricConfiguration) AsExactMatch() (ExactMatchConfiguration, errors.Error) {
	var config ExactMatchConfiguration
	err := c.as(ExactMatch, &config)
	return config, err
}

func (c MetricConfiguration) AsCompilabilityMatch() (CompilabilityMatchConfiguration, errors.Error) {
	var config CompilabilityMatchConfiguration
	err := c.as(CompilabilityMatch, &config)
	return config, err
}

func (c MetricConfiguration) as(expectedType string, destination interface{}) errors.Error {
	if val, ok := c["type"]; !ok || val != expectedType {
		return errors.New("Type Error", "Cannot interpret metric type '%s' as %s", val, expectedType)
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
	Type    string  `json:"type"`
	Measure Measure `json:"measure"`
}

type RougeSConfiguration struct {
	Type    string  `json:"type"`
	SkipN   int     `json:"skipN"`
	Measure Measure `json:"measure"`
}

type RougeNConfiguration struct {
	Type    string  `json:"type"`
	N       int     `json:"n"`
	Measure Measure `json:"measure"`
}

type BleuConfiguration struct {
	Type    string    `json:"type"`
	Weights []float64 `json:"weights"`
}

type TokenCounterConfiguration struct {
	Type string `json:"type"`
}

type ExactMatchConfiguration struct {
	Type string `json:"type"`
}

type CompilabilityMatchConfiguration struct {
	Type string `json:"type"`
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
		return errors.New("Type Error", "Cannot interpret metric type '%s' as %s", val, expectedType)
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
