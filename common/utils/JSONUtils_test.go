package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name   string `json:"name"`
	Nested struct {
		Name   string `json:"name"`
		Number int    `json:"number"`
	} `json:"nested"`
}

func TestJsonUnmarshalStrict(t *testing.T) {
	// given
	jsonStr := `{"name":"asd","nested":{"name":"nestedAsd","number":1}}`
	var output TestStruct

	// when
	UnmarshalJSONStrict([]byte(jsonStr), &output)

	// then
	assert.Equal(t, "nestedAsd", output.Nested.Name)
}

func TestJsonUnmarshalStrictWithUnknownField(t *testing.T) {
	// given
	jsonStr := `{"name":"asd","unknown":1}`
	var output TestStruct

	// when
	err := UnmarshalJSONStrict([]byte(jsonStr), &output)

	// then
	assert.Error(t, err)
}

func TestGetNestedFieldValue(t *testing.T) {
	testData := map[string]interface{}{
		"languageServer": map[string]interface{}{
			"models": map[string]interface{}{
				"methodGenerator": "someValue",
			},
		},
	}

	value, err := GetNestedValueOfMap(testData, "languageServer.models.methodGenerator")

	assert.NoError(t, err)
	assert.Equal(t, "someValue", value)
}

func TestSetNestedFieldValue(t *testing.T) {
	testData := map[string]interface{}{
		"languageServer": map[string]interface{}{
			"models": map[string]interface{}{
				"methodGenerator": "someValue",
			},
		},
	}

	err := SetNestedValueOfMap(testData, "languageServer.models", nil)

	assert.NoError(t, err)
	assert.Nil(t, testData["languageServer"].(map[string]interface{})["models"])
}

func TestDeleteNestedField(t *testing.T) {
	testData := map[string]interface{}{
		"languageServer": map[string]interface{}{
			"models": map[string]interface{}{
				"methodGenerator": "someValue",
			},
		},
	}

	err := DeleteNestedFieldOfMap(testData, "languageServer.models")

	assert.NoError(t, err)
	languageServer := testData["languageServer"].(map[string]interface{})
	_, fieldExists := languageServer["models"]
	assert.False(t, fieldExists)
}
