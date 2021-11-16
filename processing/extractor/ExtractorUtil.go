package extractor

import (
	"io/ioutil"
	"os"
	"strings"

	"returntypes-langserver/common/errors"
)

// Returns file infos for all xml files in the given directory.
func FindProjectXMLFiles(inputDir string) ([]os.FileInfo, errors.Error) {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return nil, errors.Wrap(err, "XML Error", "Could not read XML files of directory "+inputDir)
	}
	xmlfiles := make([]os.FileInfo, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".xml") {
			xmlfiles = append(xmlfiles, file)
		}
	}
	return xmlfiles, nil
}
