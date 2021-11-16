package crawler

import (
	"fmt"
	"path/filepath"
	"testing"

	"returntypes-langserver/common/configuration"

	"github.com/stretchr/testify/assert"
)

// An absolute path to the crawler .jar file. (Can not be relative as go test might build/run the program in a completely different directory)
const CrawlerPath = ""

func TestGetCodeElements(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(createCrawlerConfig())

	// when
	elements, err := GetCodeElements(getTestFilePath(), NewOptions().WithAbsolutePaths(true).Build())

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements.CodeFiles(), 1)
}

func TestGetDirectoryElements(t *testing.T) {
	// given
	configuration.LoadConfigFromJsonString(createCrawlerConfig())

	// when
	elements, err := GetCodeElementsOfDirectory(getTestFilesDir(), NewOptions().WithAbsolutePaths(true).Build())

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements.CodeFiles(), 44)
}

// Test helper functions

func getTestResourcesPath() string {
	// Need to specify an absolute path as the test build will be stored somewhere in the temp folders
	return `C:\Users\work\vscextension\returntypes-extension\test-resources`
}

func getTestFilePath() string {
	return filepath.Join(getTestResourcesPath(), "javaFiles", "Converters.java")
}

func getTestFilesDir() string {
	return filepath.Join(getTestResourcesPath(), "javaFiles")
}

func createCrawlerConfig() string {
	return fmt.Sprintf(`{"crawlerPath":"%s"}`, CrawlerPath)
}
