package predictor

import (
	"fmt"
	"testing"

	"returntypes-langserver/common/configuration"

	"github.com/stretchr/testify/assert"
)

// The tests in this file require the predictor to already run on a given host (local machine or remote)
// These constants define the address used to connect to the predictor in the tests.
const PredictorPort = 10000
const PredictorHost = "192.168.178.42"

func TestPredict(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(buildPredictorConfig())
	methods := make([]PredictableMethodName, 2)
	methods[0] = GetPredictableMethodName("getName")
	methods[1] = GetPredictableMethodName("findItem")

	// when
	Predict(methods)
	Predict(methods)
	elements, err := Predict(methods)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements, 2)
}

func TestTrain(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(buildPredictorConfig())

	// when
	evaluation, err := Train(createDataset("labels"), createDataset("training"), createDataset("evaluation"))

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
	configuration.LoadConfigFromJsonString(buildPredictorConfigForUnstableConnection())
	methods := make([]PredictableMethodName, 2)
	methods[0] = GetPredictableMethodName("getName")
	methods[1] = GetPredictableMethodName("findItem")

	// when
	Predict(methods)
	Predict(methods)
	elements, err := Predict(methods)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements, 2)
}

// Test helper functions

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
