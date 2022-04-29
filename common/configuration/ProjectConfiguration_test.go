package configuration

import (
	"encoding/json"
	"returntypes-langserver/common/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalProjectConfiguration(t *testing.T) {
	// given
	rawStr := `{"projects":["test", {"alternativeName": "someAlternativeName", "gitUri": "https://github.io/owner/repository//"}]}`
	createDefaultConfig()

	// when
	err := loadJsonConfig([]byte(rawStr))

	// then
	assert.NoError(t, err)
	assert.Len(t, loadedConfig.Projects, 2)
	assert.Equal(t, loadedConfig.Projects[0].GitUri, "test")
	assert.Equal(t, loadedConfig.Projects[1].AlternativeName, "someAlternativeName")
	assert.Equal(t, loadedConfig.Projects[1].GitUri, "https://github.io/owner/repository") // the uri should not end with a slash
}

func TestUnmarshalProjectConfigurationByMap(t *testing.T) {
	// given
	rawStr := `{"projects":["test", {"alternativeName": "someAlternativeName", "gitUri": "https://github.io/owner/repository//"}]}`
	createDefaultConfig()

	// when
	var v interface{}
	var projects ProjectConfigurationFile
	err := json.Unmarshal([]byte(rawStr), &v)
	err = utils.DecodeMapToStruct(v, &projects)

	// then
	assert.NoError(t, err)
	assert.Len(t, projects.Projects, 2)
	assert.Equal(t, projects.Projects[0].GitUri, "test")
	assert.Equal(t, projects.Projects[1].AlternativeName, "someAlternativeName")
	assert.Equal(t, projects.Projects[1].GitUri, "https://github.io/owner/repository") // the uri should not end with a slash
}
