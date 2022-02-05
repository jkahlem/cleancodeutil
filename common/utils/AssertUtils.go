package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertStringSlice(t *testing.T, actual []string, expectedValues ...string) {
	assert.Equal(t, len(expectedValues), len(actual))
	for i, expected := range expectedValues {
		assert.Equal(t, expected, actual[i])
	}
}
