package languageserver

import (
	"returntypes-langserver/common/code/java/parser"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/languageserver/workspace"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMethods(t *testing.T) {
	// given
	setupTest()
	ls := languageServer{}
	doc := workspace.NewDocument("doSomething(contentToRemove)")
	method := Method{
		Method: parser.Method{
			Name: parser.Token{
				Content: "doSomething",
			},
			RoundBraces: parser.Token{
				Content: "(contentToRemove)",
				Range: parser.Range{
					Start: 12,
					End:   27,
				},
			},
			Type: parser.Token{},
		},
	}

	// when
	items, err := ls.CompleteMethodDefinition(method, &doc)

	// then
	assert.NoError(t, err)
	if assert.Len(t, items, 1) && assert.NotNil(t, items[0].TextEdit) {
		item := items[0]
		assert.Equal(t, "Object mockParameter", item.TextEdit.NewText)
		assert.Equal(t, 13, item.TextEdit.Range.Start.Character)
		assert.Equal(t, 26, item.TextEdit.Range.End.Character)
	}
}

func setupTest() {
	config := `{
		"predictor":{
			"useMock": true
		},
		"languageServer":{
			"models":{
				"methodGenerator":"test"
			}
		},
		"datasets": [{
			"name": "test"
		}]
	}`

	configuration.MustLoadConfigFromJsonString(config)
}
