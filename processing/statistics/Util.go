package statistics

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"returntypes-langserver/common/debug/errors"
)

type FileNodesOnlyContainer struct {
	XMLName xml.Name   `xml:"root"`
	Files   []FileNode `xml:"files>file"`
}

type FileNode struct {
	XMLName  xml.Name `xml:"file"`
	FilePath string   `xml:"path,attr"`
}

// Loads only the required data (= files in the projects) from the xml file and discards anything else to safe resources
func loadOnlyFileNodesFromXML(path string) (FileNodesOnlyContainer, errors.Error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return FileNodesOnlyContainer{}, errors.Wrap(err, StatisticsErrorTitle, "Cannot read XML files")
	}
	defer xmlFile.Close()

	contents, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return FileNodesOnlyContainer{}, errors.Wrap(err, StatisticsErrorTitle, "Cannot read XML files")
	}
	nodes := FileNodesOnlyContainer{}
	if err = xml.Unmarshal(contents, &nodes); err != nil {
		return FileNodesOnlyContainer{}, errors.Wrap(err, StatisticsErrorTitle, "Cannot read XML files")
	}
	return nodes, nil
}
