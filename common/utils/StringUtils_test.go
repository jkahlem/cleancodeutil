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

func TestStringStack(t *testing.T) {
	// given
	s := NewStringStack()

	// when
	s.Push("string1")
	s.Push("string2")

	// then
	val, ok := s.Peek()
	assert.True(t, ok)
	assert.Equal(t, "string2", val)
	val, ok = s.Pop()
	assert.True(t, ok)
	assert.Equal(t, "string2", val)
	val, ok = s.Pop()
	assert.True(t, ok)
	assert.Equal(t, "string1", val)
	val, ok = s.Pop()
	assert.False(t, ok)
}
