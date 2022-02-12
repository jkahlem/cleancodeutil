package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertStringSlice(t *testing.T, actual []string, expectedValues ...string) {
	assert.Equal(t, strings.Join(expectedValues, ","), strings.Join(actual, ","))
}
