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
	method := parser.Method{
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
	}

	// when
	item, err := ls.CompleteMethodDefinition(method, &doc)

	// then
	assert.NoError(t, err)
	if assert.NotNil(t, item) && assert.NotNil(t, item.TextEdit) {
		assert.Equal(t, "Object mockParameter", item.TextEdit.NewText)
		assert.Equal(t, 13, item.TextEdit.Range.Start.Character)
		assert.Equal(t, 27, item.TextEdit.Range.End.Character)
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
