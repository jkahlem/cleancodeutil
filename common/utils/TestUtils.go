package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Asserts a string slice against the given values. Makes it easier to do exact slice assertions.
func AssertStringSlice(t *testing.T, actual []string, expectedValues ...string) {
	assert.Equal(t, strings.Join(expectedValues, ","), strings.Join(actual, ","))
}

// Gets the current working directory and panics if any error occurs.
func MustGetWd() string {
	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		return wd
	}
}
