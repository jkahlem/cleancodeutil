package java

import (
	"encoding/xml"
	"io/ioutil"
	"os"

	"returntypes-langserver/common/debug/errors"
)

const XMLErrorTitle = "XML Error"

// Opens the xml file at the path and unmarshals it.
func FromXMLFile(path string) (FileContainer, errors.Error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, XMLErrorTitle, "Could not open XML file at "+path)
	}
	defer xmlFile.Close()

	contents, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return nil, errors.Wrap(err, XMLErrorTitle, "Could not read XML file at "+path)
	}

	return UnmarshalXMLToFileContainer(contents)
}

// Unmarshals xml contents to a file container.
func UnmarshalXMLToFileContainer(contents []byte) (FileContainer, errors.Error) {
	var xmlroot XMLRoot
	if err := xml.Unmarshal(contents, &xmlroot); err != nil {
		return nil, errors.Wrap(err, XMLErrorTitle, "Could not parse XML file")
	}

	connectElements(&xmlroot)

	return &xmlroot, nil
}

// Makes sure that the unmarshalled nodes have a link to their parent.
func connectElements(root FileContainer) {
	if root == nil {
		return
	}

	visitor := ConnectorVisitor{}
	for _, file := range root.CodeFiles() {
		visitor.VisitCodeFile(file)
	}
	return
}
