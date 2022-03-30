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
	NameRaw        string         `json:"name"`
	Filter         Filter         `json:"filter"`
	IsGroupOnly    bool           `json:"isGroupOnly"`
	Description    string         `json:"description"`
	SpecialOptions SpecialOptions `json:"specialOptions"`
	ModelOptions   ModelOptions   `json:"modelOptions"`
	Subsets        []Dataset      `json:"subsets"`
	TargetModels   []string       `json:"targetModels"`
	parentPath     string
}

type SpecialOptions struct {
	Convert2ToWords     bool                    `json:"convert2ToWords"`
	MinMethodNameLength int                     `json:"minMethodNameLength"`
	FilterDuplicates    bool                    `json:"filterDuplicates"`
	TypeClasses         TypeClassConfigurations `json:"typeClasses"`
	MaxTrainingRows     int                     `json:"maxTrainingRows"`
	MaxEvaluationRows   int                     `json:"maxEvaluationRows"`
	// The size of the splitted datasets as a proportion
	DatasetSize DatasetProportion `json:"datasetSize"`
}

type ModelOptions struct {
	NumOfEpochs     int                         `json:"numOfEpochs"`
	BatchSize       int                         `json:"batchSize"`
	GenerationTasks MethodGenerationTaskOptions `json:"generationTasks"`
	// Sets the number of expected return sequences to predict different suggestions
	NumReturnSequences int `json:"numReturnSequences"`
	// Sets the maximum length of the predicted sequence
	MaxSequenceLength int `json:"maxSequenceLength"`
}

type MethodGenerationTaskOptions struct {
	// Defines, which tasks should also be performed when generating parameter names in the same task
	ParameterNames CompounTaskOptions `json:"parameterNames"`
	// If true, parameter type generation is performed in a separate task
	ParameterTypes bool `json:"parameterTypes"`
	// If true, return type generation is performed in a separate task
	ReturnType bool `json:"returnType"`
}

type CompounTaskOptions struct {
	// If true, the parameter list generation will be extended by return type generation in the same task
	WithReturnType bool `json:"withReturnType"`
	// If true, the parameter list generation will be extended by parameter type generation in the same task
	WithParameterTypes bool `json:"withParameterTypes"`
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
	}
}
