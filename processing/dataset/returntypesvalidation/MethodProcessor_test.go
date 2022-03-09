package returntypesvalidation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnTypesSet(t *testing.T) {
	// given
	set := make(ReturnTypes)

	// when
	set.Put("number")
	set.Put("number")
	set.Put("string")

	// then
	assert.Equal(t, "number", set.MostUsedType())
}
