package predictor

import (
	"sync"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
)

var singleton Predictor
var singletonMutex sync.Mutex

// Makes predictions for the methods in the map and sets the types as their value.
func PredictToMap(mapping MethodTypeMap) errors.Error {
	return getSingleton().PredictToMap(mapping)
}

// Predicts the expected return type for the given method names. Returns a list of expected return types in the exact order
// the method names were passed.
func Predict(methodNames []PredictableMethodName) ([]string, errors.Error) {
	return getSingleton().Predict(methodNames)
}

// Starts the training and evaluation process. Returns the evaluation result if finished.
func Train(labels, trainingSet, evaluationSet [][]string) (Evaluation, errors.Error) {
	return getSingleton().Train(labels, trainingSet, evaluationSet)
}

func getSingleton() Predictor {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if singleton == nil {
		singleton = createSingleton()
	}
	return singleton
}

func createSingleton() Predictor {
	if configuration.PredictorUseMock() {
		log.Info("Setup predictor using predictor mock...\n")
		return &mock{}
	}
	return &predictor{}
}
