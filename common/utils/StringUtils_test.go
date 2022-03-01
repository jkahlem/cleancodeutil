package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSet(t *testing.T) {
	// given
	s := make(StringSet)

	// when
	s.Put("test")
	s.Put("test2")

	// then
	assert.True(t, s.Has("test"))
	assert.True(t, s.Has("test2"))
	assert.False(t, s.Has("unknown"))
}
