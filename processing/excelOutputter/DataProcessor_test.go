package excelOutputter

import (
	"encoding/json"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/services/predictor"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataProcessorAccepting(t *testing.T) {
	// given
	d := Dataset{Filter: Filter{
		Includes: buildFilter(`{"method": ["test*"]}`),
	}}
	processor := NewDatasetProcessor(d, "")

	// then
	assert.True(t, processor.accepts(csv.Method{MethodName: predictor.SplitMethodNameToSentence("testMethodFilter")}))
	assert.False(t, processor.accepts(csv.Method{MethodName: predictor.SplitMethodNameToSentence("someMethodToExclude")}))
}

func buildFilter(raw string) *FilterConfiguration {
	filter := FilterConfiguration{}
	err := json.Unmarshal([]byte(raw), &filter)
	if err != nil {
		panic(err)
	}
	return &filter
}
