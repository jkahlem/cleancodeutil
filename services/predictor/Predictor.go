// The predictor package is used for communicating with the predictor application.
// The package defines a high-level API for training the predictor using datasets
// and predict the return types of given method names.
package predictor

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

const PredictorErrorTitle = "Predictor Error"

type SupportedModels string

const (
	ReturnTypesPrediction SupportedModels = "ReturnTypesPrediction"
	MethodGenerator       SupportedModels = "MethodGenerator"
)

type MethodTypeMap map[PredictableMethodName]string

// Interface used for the predictor to support multiple predictor implementations like the mock.
type Predictor interface {
	// Makes predictions for the methods in the map and sets the types as their value.
	PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error
	// Predicts the expected return type for the given method names. Returns a list of expected return types in the exact order
	// the method names were passed.
	PredictReturnTypes(methodNames []PredictableMethodName) ([]MethodValues, errors.Error)
	// Evaluates the passed methods and returns the scores for it
	EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error)
	// Generates the remained part of a method by it's method name
	GenerateMethods(contexts []MethodContext) ([]MethodValues, errors.Error)
	// Starts the training and evaluation process.
	TrainReturnTypes(methods []Method, labels [][]string) errors.Error
	// Starts the training and evaluation process.
	TrainMethods(trainingSet []Method) errors.Error
	// Returns true if the model exists and is already trained
	ModelExists(modelType SupportedModels) (bool, errors.Error)
}

type predictor struct {
	config configuration.Dataset
}

func OnDataset(dataset configuration.Dataset) Predictor {
	if configuration.PredictorUseMock() {
		return &mock{}
	}
	return &predictor{
		config: dataset,
	}
}

func (p *predictor) ModelExists(modelType SupportedModels) (bool, errors.Error) {
	options := p.getOptions(modelType)
	return remote().Exists(options)
}

func (p *predictor) TrainReturnTypes(methods []Method, labels [][]string) errors.Error {
	options := p.getOptions(ReturnTypesPrediction)
	options.LabelsCsv = p.asCsvString(labels)
	return remote().Train(methods, options)
}

func (p *predictor) EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error) {
	options := p.getOptions(ReturnTypesPrediction)
	options.LabelsCsv = p.asCsvString(labels)
	return remote().Evaluate(evaluationSet, options)
}

func (p *predictor) PredictReturnTypes(methodNames []PredictableMethodName) ([]MethodValues, errors.Error) {
	options := p.getOptions(ReturnTypesPrediction)
	contexts := make([]MethodContext, len(methodNames))
	for i, name := range methodNames {
		contexts[i].MethodName = name
	}
	return remote().Predict(contexts, options)
}

// Makes predictions for the methods in the map and sets the types as their value.
func (p *predictor) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	names := p.getMethodNamesInsideOfMap(mapping)
	predictedTypes, err := p.PredictReturnTypes(names)
	if err != nil {
		return err
	}

	if len(names) != len(predictedTypes) {
		return errors.New(PredictorErrorTitle, "Expected %d predictions, but got %d.", len(names), len(predictedTypes))
	}

	for index, name := range names {
		mapping[name] = predictedTypes[index].ReturnType
	}
	return nil
}

func (p *predictor) getMethodNamesInsideOfMap(mapping MethodTypeMap) []PredictableMethodName {
	names := make([]PredictableMethodName, len(mapping))
	i := 0
	for methodName := range mapping {
		names[i] = methodName
		i++
	}
	return names[:i]
}

func (p *predictor) TrainMethods(trainingSet []Method) errors.Error {
	return remote().Train(trainingSet, p.getOptions(MethodGenerator))
}

func (p *predictor) GenerateMethods(contexts []MethodContext) ([]MethodValues, errors.Error) {
	return remote().Predict(contexts, p.getOptions(MethodGenerator))
}

func (p *predictor) getOptions(modelType SupportedModels) Options {
	return Options{
		Identifier: p.config.QualifiedIdentifier(),
		Type:       modelType,
	}
}

func (p *predictor) asCsvString(records [][]string) string {
	builder := strings.Builder{}
	csv.NewWriter(&builder).WriteAllRecords(records)
	return builder.String()
}
