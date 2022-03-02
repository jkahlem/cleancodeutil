package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalProjectConfiguration(t *testing.T) {
	// given
	rawStr := `{"projects":["test", {"alternativeName": "someAlternativeName"}]}`
	createDefaultConfig()

	// when
	err := loadJsonConfig([]byte(rawStr))

	// then
	assert.NoError(t, err)
	assert.Len(t, loadedConfig.Projects, 2)
	assert.Equal(t, loadedConfig.Projects[0].GitUri, "test")
	assert.Equal(t, loadedConfig.Projects[1].AlternativeName, "someAlternativeName")
}
