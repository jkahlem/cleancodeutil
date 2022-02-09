package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalRequest(t *testing.T) {
	// given
	rawRequest := `
	{
		"jsonrpc":"2.0",
		"id":0,
		"method":"testRequest",
		"params":[{"testKey": "testValue"}]
	}`

	// when
	json, err := Unmarshal(rawRequest)
	request, ok := json.(Request)

	// then
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "2.0", request.JsonRPC)
	assert.Equal(t, float64(0), request.Id)
	assert.Equal(t, "testRequest", request.Method)
	assertParams(t, request.Params)
}

func TestUnmarshalSuccessfulResponse(t *testing.T) {
	// given
	rawResponse := `
	{
		"jsonrpc":"2.0",
		"id":0,
		"result":{"testKey": "testValue"}
	}`

	// when
	json, err := Unmarshal(rawResponse)
	response, ok := json.(Response)

	// then
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "2.0", response.JsonRPC)
	assert.Equal(t, float64(0), response.Id)
	assertTestObj(t, response.Result)
}

func TestUnmarshalErrorResponse(t *testing.T) {
	// given
	rawResponse := `
	{
		"jsonrpc":"2.0",
		"id":0,
		"error": {
			"code": -32700,
			"message": "TestErrorMsg"
		}
	}`

	// when
	json, err := Unmarshal(rawResponse)
	response, ok := json.(Response)

	// then
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "2.0", response.JsonRPC)
	assert.Equal(t, float64(0), response.Id)
	assertResponseError(t, ErrorCode(-32700), "TestErrorMsg", response.Error)
}

func TestUnmarshalNotification(t *testing.T) {
	// given
	rawNotification := `
	{
		"jsonrpc":"2.0",
		"method":"testNotification",
		"params":[{"testKey": "testValue"}]
	}`

	// when
	json, err := Unmarshal(rawNotification)
	notification, ok := json.(Notification)

	// then
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "2.0", notification.JsonRPC)
	assert.Equal(t, "testNotification", notification.Method)
	assertParams(t, notification.Params)
}

// Helper functions

func assertParams(t *testing.T, params interface{}) {
	paramsSlice, ok := params.([]interface{})
	assert.True(t, ok)
	assert.Len(t, paramsSlice, 1)
	assertTestObj(t, paramsSlice[0])
}

func assertTestObj(t *testing.T, obj interface{}) {
	paramsObj, ok := obj.(map[string]interface{})
	assert.True(t, ok)

	value, keyExists := paramsObj["testKey"]
	assert.True(t, keyExists)
	assert.Equal(t, "testValue", value)
}

func assertResponseError(t *testing.T, expectedErrorCode ErrorCode, expectedErrorMsg string, responseError interface{}) {
	err, ok := responseError.(*ResponseError)
	assert.True(t, ok)
	assert.NotNil(t, err)
	assert.Equal(t, expectedErrorCode, err.Code)
	assert.Equal(t, expectedErrorMsg, err.Message)
}
