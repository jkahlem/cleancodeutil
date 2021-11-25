package predictor

import (
	"returntypes-langserver/common/errors"
)

// Implements the predictor interface without using the external service for testing purposes where
// the results are not important.
//
// The mock will not call the remote service and has no other dependencies.
// The mock will always predict the type specified in MockReturnTypePrediction for any method to predict.
type mock struct{}

const MockReturnTypePrediction = "void"
const MockGeneratedMethod = ": <object> parameter"

// Maps everything to the specified type.
func (p *mock) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	for key := range mapping {
		mapping[key] = MockReturnTypePrediction
	}
	return nil
}

// Predicts for each method name the specified type.
func (p *mock) PredictReturnTypes(methodNames []PredictableMethodName) ([]string, errors.Error) {
	returnTypes := make([]string, len(methodNames))
	for i := range returnTypes {
		returnTypes[i] = MockReturnTypePrediction
	}
	return returnTypes, nil
}

// GEnerates a method
func (p *mock) GenerateMethods(methodNames []PredictableMethodName) ([]string, errors.Error) {
	returnTypes := make([]string, len(methodNames))
	for i := range returnTypes {
		returnTypes[i] = string(methodNames[i]) + MockGeneratedMethod
	}
	return returnTypes, nil
}

// Always returns a successful evaluation result.
func (p *mock) TrainReturnTypes(labels, trainingSet, evaluationSet [][]string) (Evaluation, errors.Error) {
	return Evaluation{
		AccScore: 1,
		EvalLoss: 1,
		MCC:      1,
		F1Score:  1,
	}, nil
}

// Always returns a successful evaluation result.
func (p *mock) TrainMethods(trainingSet, evaluationSet [][]string) (Evaluation, errors.Error) {
	return Evaluation{
		AccScore: 1,
		EvalLoss: 1,
		MCC:      1,
		F1Score:  1,
	}, nil
}
