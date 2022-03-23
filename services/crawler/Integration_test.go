package crawler

import (
	"fmt"
	"path/filepath"
	"testing"

	"returntypes-langserver/common/configuration"

	"github.com/stretchr/testify/assert"
)

// An absolute path to the crawler .jar file. (Can not be relative as go test might build/run the program in a completely different directory)
const CrawlerPath = `C:\\Users\\work\\Documents\\bachelor\\02-project\\returntypes-predictor\\mainapp\\resources\\crawler\\returntypes-crawler.jar`

func TestGetCodeElements(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(createCrawlerConfig())

	// when
	elements, err := GetCodeElements(getTestFilePath(), NewOptions().WithAbsolutePaths(true).Build())

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements.CodeFiles(), 1)
}

func TestGetDirectoryElements(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(createCrawlerConfig())

	// when
	elements, err := GetCodeElementsOfDirectory(getTestFilesDir(), NewOptions().WithAbsolutePaths(true).Build())

	// then
	assert.NoError(t, err)
	assert.NotNil(t, elements)
	assert.Len(t, elements.CodeFiles(), 44)
}

func TestParseSourceCodeForUnfinishedCode(t *testing.T) {
	// given
	configuration.MustLoadConfigFromJsonString(createCrawlerConfig())
	code := `package com.example;

public class Example {
	public String name;
	public String getName() {
		return name;
	}

	public void printName(

	public void setName(String name) {
		this.name = name;
	}
}`

	// when
	elements, err := ParseSourceCode(code, NewOptions().Build())

	// then
	assert.NoError(t, err)
	assert.Len(t, elements.CodeFiles()[0].Classes[0].Methods, 3)

	methods := elements.CodeFiles()[0].Classes[0].Methods
	assert.Equal(t, "getName", methods[0])
	assert.Equal(t, "printName", methods[1])
	assert.Equal(t, "setName", methods[2])
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
	return fmt.Sprintf(`{"crawler":{"executablePath":"%s"}}`, CrawlerPath)
}
