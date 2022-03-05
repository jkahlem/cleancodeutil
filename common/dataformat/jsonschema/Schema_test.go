package jsonschema

import (
	"path/filepath"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemas(t *testing.T) {
	// given
	raw := `{
		"include": {
			"method": ["asd", {
				"pattern": "a pattern",
				"type": "regexp"
			}]
		}
	}`
	root := utils.FilePathToURI(filepath.Join(utils.MustGetWd(), "testdata"))
	schema, err := AtRoot(root).
		WithTopLevel("filter.schema.json").
		WithResources("filter-configuration.schema.json",
			"pattern.schema.json").
		Compile()

	if assert.NoError(t, err) {
		// when
		err = schema.Validate(utils.MustUnmarshalJsonToMap(raw))

		// then
		assert.NoError(t, err)
	}
}
