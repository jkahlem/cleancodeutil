package configuration

import (
	"fmt"
	"os"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/utils"
)

type DatasetConfiguration []Dataset

type DatasetFileConfiguration struct {
	Datasets []Dataset `json:"datasets"`
}

type DatasetBase struct {
	NameRaw      string       `json:"name"`
	Description  string       `json:"description"`
	ModelOptions ModelOptions `json:"modelOptions"`
	TargetModels []string     `json:"targetModels"`
}

type Dataset struct {
	DatasetBase    `json:",squash"`
	Filter         Filter         `json:"filter"`
	SpecialOptions SpecialOptions `json:"specialOptions"`
	Subsets        []Dataset      `json:"subsets"`
	Alternatives   []DatasetBase  `json:"alternatives"`
	parentPath     string
}

type SpecialOptions struct {
	MaxTokensPerOutputSequence int                     `json:"maxTokensPerOutputSequence"`
	FilterDuplicates           bool                    `json:"filterDuplicates"`
	TypeClasses                TypeClassConfigurations `json:"typeClasses"`
	MaxTrainingRows            int                     `json:"maxTrainingRows"`
	MaxEvaluationRows          int                     `json:"maxEvaluationRows"`
	// The size of the splitted datasets as a proportion
	DatasetSize        DatasetProportion         `json:"datasetSize"`
	SentenceFormatting SentenceFormattingOptions `json:"sentenceFormatting"`
}

type ModelOptions struct {
	NumOfEpochs     int                          `json:"numOfEpochs"`
	BatchSize       int                          `json:"batchSize"`
	GenerationTasks *MethodGenerationTaskOptions `json:"generationTasks,omitempty"`
	// Sets the number of expected return sequences to predict different suggestions
	NumReturnSequences int `json:"numReturnSequences"`
	// Sets the maximum length of the predicted sequence
	MaxSequenceLength int `json:"maxSequenceLength"`
	// If true, the types available in a file are also sent to the predictor.
	UseContextTypes bool `json:"useContextTypes"`
	// If true, a keyword (like 'void') is used to indicate no output.
	EmptyParameterListByKeyword bool `json:"emptyParameterListByKeyword"`
	// Adafactor optimizer options
	Adafactor Adafactor `json:"adafactor"`
	// Defines the model to use for the task
	ModelName     string   `json:"modelName"`
	NumBeams      int      `json:"numBeams"`
	LengthPenalty *float64 `json:"lengthPenalty,omitempty"`
	TopK          *float64 `json:"topK,omitempty"`
	TopN          *float64 `json:"topN,omitempty"`
	OutputOrder   []string `json:"outputOrder"`
}

type Adafactor struct {
	Beta           *float64  `json:"beta,omitempty"`
	ClipThreshold  *float64  `json:"clipThreshold,omitempty"`
	DecayRate      *float64  `json:"decayRate,omitempty"`
	Eps            []float64 `json:"eps,omitempty"`
	RelativeStep   *bool     `json:"relativeStep,omitempty"`
	WarmupInit     *bool     `json:"warmupInit,omitempty"`
	ScaleParameter *bool     `json:"scaleParameter,omitempty"`
}

type MethodGenerationTaskOptions struct {
	// Defines, which tasks should also be performed when generating parameter names in the same task
	ParameterNames CompoundTaskOptions `json:"parameterNames"`
	// If true, parameter type generation is performed in a separate task
	ParameterTypes bool `json:"parameterTypes"`
	// If true, return type generation is performed in a separate task
	ReturnType bool `json:"returnType"`
}

type CompoundTaskOptions struct {
	// If true, the parameter list generation will be extended by return type generation in the same task
	WithReturnType bool `json:"withReturnType"`
	// If true, the parameter list generation will be extended by parameter type generation in the same task
	WithParameterTypes bool `json:"withParameterTypes"`
}

type SentenceFormattingOptions struct {
	// If true, method names should be splitted for the model/evaluation.
	MethodName bool `json:"methodName"`
	// If true, type names should be splitted for the model/evaluation.
	TypeName bool `json:"typeName"`
	// If true, parameter names should be splitted for the model/evaluation.
	ParameterName bool `json:"parameterName"`
}

func (o SentenceFormattingOptions) DecodeValue(value interface{}) (interface{}, error) {
	if boolVal, ok := value.(bool); ok {
		return SentenceFormattingOptions{
			MethodName:    boolVal,
			TypeName:      boolVal,
			ParameterName: boolVal,
		}, nil
	}
	return value, nil
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
	if filePath == "" {
		return nil
	}
	contents, err := os.ReadFile(AbsolutePathFromGoProjectDir(filePath))
	if err != nil {
		return err
	}
	return c.fromJson(contents)
}

func (c *DatasetConfiguration) fromJson(contents []byte) error {
	var config DatasetFileConfiguration
	if err := jsonschema.UnmarshalJSONStrict(contents, &config, DatasetConfigurationFileSchema); err != nil {
		return err
	}
	*c = config.Datasets
	return nil
}

// Returns the qualified identifier of the dataset
func (c *Dataset) QualifiedIdentifier() string {
	if c.parentPath != "" {
		return c.parentPath + "/" + c.Name()
	}
	return c.Name()
}

func (c *Dataset) Name() string {
	return DatasetPrefix() + c.NameRaw
}

func connectDatasetPaths(datasets []Dataset, parentPath string) {
	for i := range datasets {
		datasets[i].parentPath = parentPath
		connectDatasetPaths(datasets[i].Subsets, datasets[i].QualifiedIdentifier())
		for j, alt := range datasets[i].Alternatives {
			datasets[i].Alternatives[j].NameRaw = fmt.Sprintf("%s_%s", datasets[i].NameRaw, alt.NameRaw)
		}
	}
}

func (c Dataset) DecodeValue(value interface{}) (interface{}, error) {
	// Before mapping the json output (map[string]interface{}) to a dataset struct,
	// merge the modelOptions/specialOptions of this dataset to the modelOptions/specialOptions
	// of each subset (so set only keys which are unset).
	//
	// This approach allows leaving the implementation of the data structure as it currently is
	// without implementing pointers and pointer checks everywhere as otherwise it is unknown
	// if a value (bool/int values etc.) is explicitly set to a zero value or was just omitted.
	if jsonObj, ok := value.(map[string]interface{}); ok {
		sourceModelOptions, hasModelOptions := jsonObj["modelOptions"]
		sourceSpecialOptions, hasSpecialOptions := jsonObj["specialOptions"]

		if subsets, ok := jsonObj["subsets"]; ok {
			if subsetSlice, ok := subsets.([]interface{}); ok {
				if hasModelOptions {
					if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
						c.mergeModelOptionsToSubsets(typedModelOptions, subsetSlice)
					}
				}
				if hasSpecialOptions {
					if typedSpecialOptions, ok := sourceSpecialOptions.(map[string]interface{}); ok {
						c.mergeSpecialOptionsToSubsets(typedSpecialOptions, subsetSlice)
					}
				}
			}
		}
		if alternatives, ok := jsonObj["alternatives"]; ok {
			if alternativesSlice, ok := alternatives.([]interface{}); ok {
				if hasModelOptions {
					if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
						c.mergeModelOptionsToSubsets(typedModelOptions, alternativesSlice)
					}
				}
			}
		}
	}
	return value, nil
}

func (c Dataset) mergeOptionsToSubsets(jsonObj map[string]interface{}, subsets []interface{}) {
	sourceModelOptions, hasModelOptions := jsonObj["modelOptions"]
	sourceSpecialOptions, hasSpecialOptions := jsonObj["specialOptions"]

	if subsets, ok := jsonObj["subsets"]; ok {
		if subsetSlice, ok := subsets.([]interface{}); ok {
			if hasModelOptions {
				if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
					c.mergeModelOptionsToSubsets(typedModelOptions, subsetSlice)
				}
			}
			if hasSpecialOptions {
				if typedSpecialOptions, ok := sourceSpecialOptions.(map[string]interface{}); ok {
					c.mergeSpecialOptionsToSubsets(typedSpecialOptions, subsetSlice)
				}
			}
		}
	}
}

func (c Dataset) mergeModelOptionsToSubsets(sourceModelOptions map[string]interface{}, subsets []interface{}) {
	for i := range subsets {
		if subsetObject, ok := subsets[i].(map[string]interface{}); ok {
			if destinationModelOptions, ok := subsetObject["modelOptions"]; ok {
				if typed, ok := destinationModelOptions.(map[string]interface{}); ok {
					utils.AddUnsettedValues(sourceModelOptions, typed)
				}
			} else {
				subsetObject["modelOptions"] = sourceModelOptions
			}
		}
	}
}

func (c Dataset) mergeSpecialOptionsToSubsets(sourceSpecialOptions map[string]interface{}, subsets []interface{}) {
	for i := range subsets {
		if subsetObject, ok := subsets[i].(map[string]interface{}); ok {
			if specialOptions, ok := subsetObject["specialOptions"]; ok {
				if typed, ok := specialOptions.(map[string]interface{}); ok {
					utils.AddUnsettedValues(sourceSpecialOptions, typed)
				}
			} else {
				subsetObject["specialOptions"] = sourceSpecialOptions
			}
		}
	}
}
