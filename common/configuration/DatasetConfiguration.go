package configuration

import (
	"fmt"
	"os"
	"returntypes-langserver/common/dataformat/jsonschema"
	"returntypes-langserver/common/utils"
)

type EvaluationTarget string

const (
	BestModel EvaluationTarget = "best-model"
	Step      EvaluationTarget = "step"
	Epoch     EvaluationTarget = "epoch"
)

type DatasetConfiguration []Dataset

type DatasetFileConfiguration struct {
	Datasets []Dataset `json:"datasets"`
}

type DatasetBase struct {
	NameRaw              string               `json:"name"`
	Description          string               `json:"description,omitempty"`
	ModelOptions         ModelOptions         `json:"modelOptions"`
	TargetModels         []string             `json:"targetModels,omitempty"`
	EvaluateOn           EvaluationTarget     `json:"evaluateOn"`
	PreprocessingOptions PreprocessingOptions `json:"preprocessingOptions"`
}

type Dataset struct {
	DatasetBase     `json:",squash"`
	Filter          Filter                 `json:"filter,omitempty"`
	CreationOptions DatasetCreationOptions `json:"creationOptions"`
	Subsets         []Dataset              `json:"subsets,omitempty"`
	Alternatives    []DatasetBase          `json:"alternatives,omitempty"`
	parentPath      string
}

type DatasetCreationOptions struct {
	MaxTokensPerOutputSequence int                     `json:"maxTokensPerOutputSequence,omitempty"`
	FilterDuplicates           bool                    `json:"filterDuplicates,omitempty"`
	TypeClasses                TypeClassConfigurations `json:"typeClasses,omitempty"`
	DatasetSize                DatasetProportion       `json:"datasetSize,omitempty"`
}

type PreprocessingOptions struct {
	MaxTrainingRows    int                       `json:"maxTrainingRows,omitempty"`
	MaxEvaluationRows  int                       `json:"maxEvaluationRows,omitempty"`
	SentenceFormatting SentenceFormattingOptions `json:"sentenceFormatting"`
}

type ModelOptions struct {
	ModelType                   string    `json:"modelType"`
	ModelName                   string    `json:"modelName"`
	NumOfEpochs                 int       `json:"numOfEpochs,omitempty"`
	BatchSize                   int       `json:"batchSize,omitempty"`
	NumReturnSequences          int       `json:"numReturnSequences,omitempty"`
	MaxSequenceLength           int       `json:"maxSequenceLength,omitempty"`
	UseContextTypes             bool      `json:"useContextTypes,omitempty"`
	EmptyParameterListByKeyword bool      `json:"emptyParameterListByKeyword,omitempty"`
	Adafactor                   Adafactor `json:"adafactor"`
	Adam                        Adam      `json:"adam"`
	NumBeams                    int       `json:"numBeams,omitempty"`
	LengthPenalty               *float64  `json:"lengthPenalty,omitempty"`
	TopK                        *float64  `json:"topK,omitempty"`
	TopN                        *float64  `json:"topN,omitempty"`
	OutputOrder                 []string  `json:"outputOrder,omitempty"`
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

type Adam struct {
	LearningRate *float64 `json:"learningRate,omitempty"`
	Eps          *float64 `json:"eps,omitempty"`
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

const (
	ModelOptionsFieldName         = "modelOptions"
	CreationOptionsFieldName      = "creationOptions"
	PreprocessingOptionsFieldName = "preprocessingOptions"
)

func (c Dataset) DecodeValue(value interface{}) (interface{}, error) {
	// Before mapping the json output (map[string]interface{}) to a dataset struct,
	// merge the modelOptions/specialOptions of this dataset to the modelOptions/specialOptions
	// of each subset (so set only keys which are unset).
	//
	// This approach allows leaving the implementation of the data structure as it currently is
	// without implementing pointers and pointer checks everywhere as otherwise it is unknown
	// if a value (bool/int values etc.) is explicitly set to a zero value or was just omitted.
	if jsonObj, ok := value.(map[string]interface{}); ok {
		sourceModelOptions, hasModelOptions := jsonObj[ModelOptionsFieldName]
		sourceSpecialOptions, hasSpecialOptions := jsonObj[CreationOptionsFieldName]
		sourcePreprocessingOptions, hasPreprocessingOptions := jsonObj[PreprocessingOptionsFieldName]

		if subsets, ok := jsonObj["subsets"]; ok {
			if subsetSlice, ok := subsets.([]interface{}); ok {
				if hasModelOptions {
					if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
						c.mergeValueToSubsets(ModelOptionsFieldName, typedModelOptions, subsetSlice)
					}
				}
				if hasSpecialOptions {
					if typedSpecialOptions, ok := sourceSpecialOptions.(map[string]interface{}); ok {
						c.mergeValueToSubsets(CreationOptionsFieldName, typedSpecialOptions, subsetSlice)
					}
				}
				if hasPreprocessingOptions {
					if typedPreprocessingOptions, ok := sourcePreprocessingOptions.(map[string]interface{}); ok {
						c.mergeValueToSubsets(PreprocessingOptionsFieldName, typedPreprocessingOptions, subsetSlice)
					}
				}
			}
		}
		if alternatives, ok := jsonObj["alternatives"]; ok {
			if alternativesSlice, ok := alternatives.([]interface{}); ok {
				if hasModelOptions {
					if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
						c.mergeValueToSubsets(ModelOptionsFieldName, typedModelOptions, alternativesSlice)
					}
				}
				if hasPreprocessingOptions {
					if typedPreprocessingOptions, ok := sourcePreprocessingOptions.(map[string]interface{}); ok {
						c.mergeValueToSubsets(PreprocessingOptionsFieldName, typedPreprocessingOptions, alternativesSlice)
					}
				}
			}
		}
	}
	return value, nil
}

func (c Dataset) mergeOptionsToSubsets(jsonObj map[string]interface{}, subsets []interface{}) {
	sourceModelOptions, hasModelOptions := jsonObj[ModelOptionsFieldName]
	sourceSpecialOptions, hasSpecialOptions := jsonObj[CreationOptionsFieldName]
	sourcePreprocessingOptions, hasPreprocessingOptions := jsonObj[PreprocessingOptionsFieldName]

	if subsets, ok := jsonObj["subsets"]; ok {
		if subsetSlice, ok := subsets.([]interface{}); ok {
			if hasModelOptions {
				if typedModelOptions, ok := sourceModelOptions.(map[string]interface{}); ok {
					c.mergeValueToSubsets(ModelOptionsFieldName, typedModelOptions, subsetSlice)
				}
			}
			if hasSpecialOptions {
				if typedSpecialOptions, ok := sourceSpecialOptions.(map[string]interface{}); ok {
					c.mergeValueToSubsets(CreationOptionsFieldName, typedSpecialOptions, subsetSlice)
				}
			}
			if hasPreprocessingOptions {
				if typedPreprocessingOptions, ok := sourcePreprocessingOptions.(map[string]interface{}); ok {
					c.mergeValueToSubsets(PreprocessingOptionsFieldName, typedPreprocessingOptions, subsetSlice)
				}
			}
		}
	}
}

func (c Dataset) mergeValueToSubsets(fieldName string, sourceValue map[string]interface{}, subsets []interface{}) {
	for i := range subsets {
		if subsetObject, ok := subsets[i].(map[string]interface{}); ok {
			if fieldValue, ok := subsetObject[fieldName]; ok {
				if typed, ok := fieldValue.(map[string]interface{}); ok {
					utils.AddUnsettedValues(sourceValue, typed)
				}
			} else {
				subsetObject[fieldName] = sourceValue
			}
		}
	}
}
