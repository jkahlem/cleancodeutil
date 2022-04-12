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
		ClassName:  "testclass",
	}, Values: MethodValues{Parameters: []Parameter{
		{Name: "test", Type: "int"},
	}}}}

	// when
	err := OnDataset(dataset()).TrainMethods(trainingSet)

	// then
	assert.NoError(t, err)
}

func TestGenerateMethods(t *testing.T) {
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())

	values, err := OnDataset(configuration.Dataset{
		DatasetBase: configuration.DatasetBase{
			NameRaw: "draft-dataset-220411",
			ModelOptions: configuration.ModelOptions{
				GenerationTasks: &configuration.MethodGenerationTaskOptions{
					ParameterNames: configuration.CompoundTaskOptions{
						WithReturnType:     true,
						WithParameterTypes: true,
					},
				},
				NumOfEpochs:                 2,
				UseContextTypes:             false,
				UseTypePrefixing:            false,
				EmptyParameterListByKeyword: true,
				NumReturnSequences:          1,
			},
		},
		SpecialOptions: configuration.SpecialOptions{
			SentenceFormatting: configuration.SentenceFormattingOptions{
				MethodName:    true,
				ParameterName: true,
			},
		},
	}).GenerateMethods([]MethodContext{context("ListItem", "compareTo"),
		context("VehicleList", "forEach"),
		context("Dialog", "createWarning"),
		context("Dialog", "forException"),
		context("Exception", "build"),
		context("Exception", "withMessage"),
	})

	if assert.NoError(t, err) {
		fmt.Println(values)
	}
}

func context(class, name string) MethodContext {
	return MethodContext{
		MethodName: strings.ToLower(SplitMethodNameToSentence(name)),
		ClassName:  class,
		IsStatic:   false,
	}
}
func static(class, name string) MethodContext {
	return MethodContext{
		MethodName: strings.ToLower(SplitMethodNameToSentence(name)),
		ClassName:  class,
		IsStatic:   true,
	}
}

/*
func TestGenerateMethods(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(buildPredictorConfig())
	methods := make([]PredictableMethodName, 8)
	//methods[0] = PredictableMethodName(string(GetPredictableMethodName("findByNameOrAge")))
	//methods[1] = PredictableMethodName(string(GetPredictableMethodName("compare")))
	methods[0] = createTypeAssignmentTest("find user by name", "user name", "string, int, object, person, user")
	// Expect: "string"
	// Prediction:
	// - Simple version: not tested (with the user stuff ... otherwise this was also string.)
	// - Method name version: "string"
	methods[1] = createTypeAssignmentTest("set age", "age", "string, int, object, person")
	// Expect: "int"
	// Prediction:
	// - Simple version: "int"
	// - Method name version: "int"
	methods[2] = createTypeAssignmentTest("set active state", "state", "string, int, object, boolean")
	// Expect: "int"
	// Prediction:
	// - Simple version: "int"
	// - Method name version: "int"
	methods[3] = createTypeAssignmentTest("compare strings", "a", "string, int, object, boolean")
	// Expect: "string"
	// Prediction:
	// - Simple version: "int"
	// - Method name version: "string"
	methods[4] = createTypeAssignmentTest("compare numbers", "b", "string, int, object, boolean")
	// Expect: "int"
	// Prediction:
	// - Simple version: "int"
	// - Method name version: "int"
	methods[5] = createTypeAssignmentTest("create contract by offer", "offer", "string, int, object, boolean, sales contract, sales offer")
	// Expect: "sales offer"
	// Prediction:
	// - Simple version: "salescontract"
	// - Method name version: "salescontract"
	methods[6] = createTypeAssignmentTest("find contract", "contract id", "string, int, object, boolean, sales contract, sales offer, person")
	// Expect: "sales offer"
	// Prediction:
	// - Simple version: "salescontract"
	// - Method name version: "salescontract"
	methods[7] = createTypeAssignmentTest("set person", "person", "string, int, object, boolean, sales contract, sales offer, person")
	// Expect: "person"
	// Prediction:
	// - Simple version: not tested
	// - Method name version: "salescontract"

	// the problem with the "salescontract" as predict might be related to the fact, that there is not that much of training data and the training data has no real context.
	// Another thing which might be interesting to test, is how writing type names together (without splitting between words) affects the output.

	// when
	elements, err := GenerateMethods(methods)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements, 2)
}*/

func createTypeAssignmentTest(methodName, parameterName, context string) PredictableMethodName {
	//return PredictableMethodName(fmt.Sprintf("name: %s context: %s", parameterName, context))
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
