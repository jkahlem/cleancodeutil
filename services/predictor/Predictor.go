// The predictor package is used for communicating with the predictor application.
// The package defines a high-level API for training the predictor using datasets
// and predict the return types of given method names.
package predictor

import (
	"fmt"
	"strings"

	"returntypes-langserver/common/csv"
	"returntypes-langserver/common/errors"
)

const PredictorErrorTitle = "Predictor Error"

type SupportedModels string

const (
	ReturnTypesPrediction SupportedModels = "ReturnTypesPrediction"
	MethodGenerator       SupportedModels = "MethodGenerator"
)

// Interface used for the predictor to support multiple predictor implementations like the mock.
type Predictor interface {
	PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error
	PredictReturnTypes(methodNames []PredictableMethodName) ([]string, errors.Error)
	GenerateMethods(methodNames []PredictableMethodName) ([]string, errors.Error)
	TrainReturnTypes(labels, trainingSet, evaluationSet [][]string) (Evaluation, errors.Error)
	TrainMethods(trainingSet, evaluationSet [][]string) (Evaluation, errors.Error)
}

type predictor struct{}

// Makes predictions for the methods in the map and sets the types as their value.
func (p *predictor) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	names := p.getMethodNamesInsideOfMap(mapping)
	predictedTypes, err := p.PredictReturnTypes(names)
	if err != nil {
		return err
	}

	if len(names) != len(predictedTypes) {
		return errors.New(PredictorErrorTitle, fmt.Sprintf("Expected %d predictions, but got %d.", len(names), len(predictedTypes)))
	}

	for index, name := range names {
		mapping[name] = predictedTypes[index]
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

// Predicts the expected return type for the given method names. Returns a list of expected return types in the exact order
// the method names were passed.
func (p *predictor) PredictReturnTypes(methodNames []PredictableMethodName) ([]string, errors.Error) {
	strSlice := make([]string, len(methodNames))
	for i, name := range methodNames {
		strSlice[i] = string(name)
	}
	return remote().Predict(strSlice, ReturnTypesPrediction)
}

// Predicts the expected return type for the given method names. Returns a list of expected return types in the exact order
// the method names were passed.
func (p *predictor) GenerateMethods(methodNames []PredictableMethodName) ([]string, errors.Error) {
	strSlice := make([]string, len(methodNames))
	for i, name := range methodNames {
		strSlice[i] = string(name)
	}
	return remote().Predict(strSlice, MethodGenerator)
}

// Starts the training and evaluation process. Returns the evaluation result if finished.
func (p *predictor) TrainReturnTypes(labels, trainingSet, evaluationSet [][]string) (Evaluation, errors.Error) {
	return remote().Train(p.asCsvString(trainingSet), p.asCsvString(evaluationSet), p.asCsvString(labels), ReturnTypesPrediction)
}

// Starts training + evaluation process for method generation
func (p *predictor) TrainMethods(trainingSet, evaluationSet [][]string) (Evaluation, errors.Error) {
	return remote().Train(p.asCsvString(trainingSet), p.asCsvString(evaluationSet), "", MethodGenerator)
}

func (p *predictor) asCsvString(records [][]string) string {
	builder := strings.Builder{}
	csv.WriteRecordsToTarget(&builder, records)
	return builder.String()
}
