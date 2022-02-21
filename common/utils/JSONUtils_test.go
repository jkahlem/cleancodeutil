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
