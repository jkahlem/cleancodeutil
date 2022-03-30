package predictor

import (
	"returntypes-langserver/common/debug/errors"
)

// Implements the predictor interface without using the external service for testing purposes where
// the results are not important.
//
// The mock will not call the remote service and has no other dependencies.
// The mock will always predict the type specified in MockReturnTypePrediction for any method to predict.
type mock struct{}

const MockReturnTypePrediction = "void"

var MockParameter Parameter = Parameter{
	Name: "mockParameter",
	Type: "Object",
}

func (p *mock) TrainReturnTypes(methods []Method, labels [][]string) errors.Error {
	return nil
}

func (p *mock) EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error) {
	return Evaluation{
		AccScore: 1,
		EvalLoss: 1,
		MCC:      1,
		F1Score:  1,
	}, nil
}

func (p *mock) PredictReturnTypes(methodNames []PredictableMethodName) ([]MethodValues, errors.Error) {
	returnTypes := make([]MethodValues, len(methodNames))
	for i := range returnTypes {
		returnTypes[i].ReturnType = MockReturnTypePrediction
	}
	return returnTypes, nil
}

// Makes predictions for the methods in the map and sets the types as their value.
func (p *mock) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	for key := range mapping {
		mapping[key] = MockReturnTypePrediction
	}
	return nil
}

func (p *mock) getMethodNamesInsideOfMap(mapping MethodTypeMap) []PredictableMethodName {
	names := make([]PredictableMethodName, len(mapping))
	i := 0
	for methodName := range mapping {
		names[i] = methodName
		i++
	}
	return names[:i]
}

func (p *mock) TrainMethods(trainingSet []Method) errors.Error {
	return nil
}

func (p *mock) GenerateMethods(contexts []MethodContext) ([][]MethodValues, errors.Error) {
	methods := make([][]MethodValues, len(contexts))
	for i := range methods {
		method := MethodValues{
			Parameters: []Parameter{MockParameter},
			ReturnType: "void",
		}
		methods[i] = []MethodValues{method}
	}
	return methods, nil
}

func (p *mock) ModelExists(modelType SupportedModels) (bool, errors.Error) {
	return true, nil
}
