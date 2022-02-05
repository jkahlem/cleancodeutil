package csv

import (
	"reflect"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	SimpleText string
	List       []string
	Number     int
}

func TestReflectUnmarshalling(t *testing.T) {
	// given
	testRecord := []string{"simple text", "list,with,four,elements", "12345"}

	// when
	unmarshalled, err := unmarshal(testRecord, reflect.TypeOf(TestStruct{}))
	casted := TestStruct{}
	if c, ok := (unmarshalled.Interface()).(TestStruct); ok {
		casted = c
	}

	// then
	assert.NoError(t, err)
	assert.Equal(t, "simple text", casted.SimpleText)
	utils.AssertStringSlice(t, casted.List, "list", "with", "four", "elements")
	assert.Equal(t, 12345, casted.Number)
}

func TestReflectMarshalling(t *testing.T) {
	// given
	testStruct := TestStruct{
		SimpleText: "simple text",
		List:       []string{"list", "with", "four", "elements"},
		Number:     12345,
	}

	// when
	record, err := marshal(reflect.ValueOf(testStruct))

	// then
	assert.NoError(t, err)
	assert.Equal(t, "simple text", record[0])
	assert.Equal(t, "list,with,four,elements", record[1])
	assert.Equal(t, "12345", record[2])
}
