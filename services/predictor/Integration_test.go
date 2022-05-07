package predictor

import (
	"fmt"
	"strings"
	"testing"

	"returntypes-langserver/common/configuration"

	"github.com/stretchr/testify/assert"
)

// The tests in this file require the predictor to already run on a given host (local machine or remote)
// These constants define the address used to connect to the predictor in the tests.
const PredictorPort = 10000
const PredictorHost = "192.168.178.42"

func TestPredictReturnTypes(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())
	methods := make([]PredictableMethodName, 2)
	methods[0] = GetPredictableMethodName("getName")
	methods[1] = GetPredictableMethodName("findItem")
	predictor := OnDataset(dataset())

	// when
	predictor.PredictReturnTypes(methods)
	predictor.PredictReturnTypes(methods)
	elements, err := predictor.PredictReturnTypes(methods)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements, 2)
}

func TestTrainMethods(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())
	trainingSet := []Method{{Context: MethodContext{
		MethodName: "training",
		ClassName:  []string{"testclass"},
	}, Values: MethodValues{Parameters: []Parameter{
		{Name: "test", Type: "int"},
	}}}}

	// when
	err := OnDataset(dataset()).TrainMethods(trainingSet, false)

	// then
	assert.NoError(t, err)
}

func context(class, name string) MethodContext {
	return MethodContext{
		MethodName: strings.ToLower(SplitMethodNameToSentence(name)),
		ClassName:  []string{class},
		IsStatic:   false,
	}
}
func static(class, name string) MethodContext {
	return MethodContext{
		MethodName: strings.ToLower(SplitMethodNameToSentence(name)),
		ClassName:  []string{class},
		IsStatic:   true,
	}
}

func createTypeAssignmentTest(methodName, parameterName, context string) PredictableMethodName {
	return PredictableMethodName(fmt.Sprintf("method: %s. name: %s. context: %s.", methodName, parameterName, context))
}

func TestTrainReturnTypes(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())
	trainingSet := []Method{{Context: MethodContext{MethodName: "training"}, Values: MethodValues{ReturnType: "test"}}}

	// when
	err := OnDataset(dataset()).TrainReturnTypes(trainingSet, createDataset("labels"))

	// then
	assert.NoError(t, err)
}

func TestEvaluateReturnTypes(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())
	evaluationSet := []Method{{Context: MethodContext{MethodName: "evaluation"}, Values: MethodValues{ReturnType: "test"}}}

	// when
	evaluation, err := OnDataset(dataset()).EvaluateReturnTypes(evaluationSet, createDataset("labels"))

	// then
	assert.NoError(t, err)
	assert.Equal(t, 0.1, evaluation.AccScore)
	assert.Equal(t, 0.2, evaluation.EvalLoss)
	assert.Equal(t, 0.3, evaluation.F1Score)
	assert.Equal(t, 0.4, evaluation.MCC)
}

func TestPredictUnstable(t *testing.T) {
	// This test is not "a full automated test" as it requires to manually activate the predictor during the test.
	// It was more of a utility while debugging a connection bug.

	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfigForUnstableConnection())
	methods := make([]PredictableMethodName, 2)
	methods[0] = GetPredictableMethodName("getName")
	methods[1] = GetPredictableMethodName("findItem")
	predictor := OnDataset(dataset())

	// when
	predictor.PredictReturnTypes(methods)
	predictor.PredictReturnTypes(methods)
	elements, err := predictor.PredictReturnTypes(methods)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements, 2)
}

func TestGetModels(t *testing.T) {
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())

	values, err := Global().GetModels(MethodGenerator)

	if assert.NoError(t, err) {
		fmt.Println(values)
	}
}

// Test helper functions

func dataset() configuration.Dataset {
	return configuration.Dataset{
		DatasetBase: configuration.DatasetBase{
			NameRaw: "test",
		},
	}
}

func createDataset(text string) [][]string {
	set := make([][]string, 1)
	set[0] = make([]string, 1)
	set[0][0] = text
	return set
}

func buildPredictorConfig() string {
	return fmt.Sprintf(`{"predictor":{"host": "%s","port": %d}}`, PredictorHost, PredictorPort)
}

func buildPredictorConfigForUnstableConnection() string {
	return fmt.Sprintf(`{"predictor":{"host": "%s","port": %d}, "connection":{"reconnectionAttempts":10,"timeout":5000}}`, PredictorHost, PredictorPort)
}
