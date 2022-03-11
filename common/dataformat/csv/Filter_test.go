package csv

import (
	"returntypes-langserver/common/configuration"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	filter := configuration.Filter{
		Includes: configuration.FilterConfigurations{{
			Method: []configuration.Pattern{{
				Pattern: "from*",
				Type:    configuration.Wildcard,
			}},
		}},
	}
	includedMethod := Method{
		MethodName: "from json object",
	}
	excludedMethod := Method{
		MethodName: "get json object",
	}

	assert.True(t, IsMethodIncluded(includedMethod, filter))
	assert.False(t, IsMethodIncluded(excludedMethod, filter))
}

func TestEmptyFilter(t *testing.T) {
	filter := configuration.Filter{}
	includedMethod := Method{
		MethodName: "from json object",
	}
	excludedMethod := Method{
		MethodName: "get json object",
	}

	assert.True(t, IsMethodIncluded(includedMethod, filter))
	assert.True(t, IsMethodIncluded(excludedMethod, filter))
}
