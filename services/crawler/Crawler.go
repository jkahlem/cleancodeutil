// The crawler package is used for communicating with the crawler application.
// The package defines a high-level API for getting the contents of a java file/project
// in a java.FileContainer structure or in the raw XML format.
package crawler

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/debug/errors"
)

const CrawlerErrorTitle = "Crawler Error"

type crawler struct{} // @ServiceGenerator:ServiceDefinition

// Gets the content of one java file.
func (c *crawler) GetCodeElements(path string, options Options) (java.FileContainer, errors.Error) {
	StartProgress(options)
	defer FinishProgress()

	xml, err := remote().GetFileContent(path, options)
	if err != nil {
		return nil, err
	}
	return c.decodeXmlContent(xml)
}

// Gets the content of all java files in the specified directory.
func (c *crawler) GetCodeElementsOfDirectory(path string, options Options) (java.FileContainer, errors.Error) {
	StartProgress(options)
	defer FinishProgress()

	xml, err := remote().GetDirectoryContents(path, options)
	if err != nil {
		return nil, err
	}
	return c.decodeXmlContent(xml)
}

// Gets the content of all java files in the specified directory.
func (c *crawler) GetRawCodeElementsOfDirectory(path string, options Options) (string, errors.Error) {
	StartProgress(options)
	defer FinishProgress()

	xml, err := remote().GetDirectoryContents(path, options)
	if err != nil {
		return "", err
	}
	return xml, nil
}

func (c *crawler) ParseSourceCode(code string, options Options) (java.FileContainer, errors.Error) {
	StartProgress(options)
	defer FinishProgress()

	xml, err := remote().ParseSourceCode(code, options)
	if err != nil {
		return nil, err
	}
	return c.decodeXmlContent(xml)
}

func (c *crawler) decodeXmlContent(xml string) (java.FileContainer, errors.Error) {
	return java.UnmarshalXMLToFileContainer([]byte(xml))
}
