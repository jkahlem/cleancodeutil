package utils

import (
	"os"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
)

// Creates a file and all directories on the given path
func CreateFile(path string) (*os.File, errors.Error) {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, errors.Wrap(err, "IO Error", "Could not create file")
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, errors.Wrap(err, "IO Error", "Could not create file")
	}
	return file, nil
}
