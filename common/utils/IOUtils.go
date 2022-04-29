package utils

import (
	"bufio"
	"fmt"
	"net/url"
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

// Writes the prompt to stdout and waits for user input. Returns the line of the user input.
func PromptUser(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(prompt)
	text, _ := reader.ReadString('\n')
	return text
}

// Returns true if the file on the given path exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Returns true if the file on the given path exists.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Returns the filepath as a DocumentURI in the file scheme.
func FilePathToURI(path string) string {
	return fmt.Sprintf("file:///%s", filepath.ToSlash(path))
}

// Parses a DocumentURI as a local filepath if possible, otherwise returns an error.
func URIToFilePath(uri string) (string, errors.Error) {
	fileUrl, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", errors.Wrap(err, "Error", "Could not parse URI")
	} else if fileUrl.Scheme != "file" {
		return "", errors.New("Error", "File scheme not supported")
	}

	path := fileUrl.Path[1:]
	return filepath.FromSlash(path), nil
}

type LazyFile struct {
	path string
	fd   *os.File
}

// Opens a file with the given path lazily, therefore the file is actually only opened/created when performing writing/reading operations.
func OpenFileLazy(path string) *LazyFile {
	return &LazyFile{
		path: path,
	}
}

func (f *LazyFile) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	} else if f.fd == nil {
		if err := f.openOrCreateFile(); err != nil {
			return 0, err
		}
	}
	return f.fd.Write(p)
}

func (f *LazyFile) openOrCreateFile() error {
	if file, err := CreateFile(f.path); err != nil {
		return err
	} else {
		f.fd = file
		return nil
	}
}

func (f *LazyFile) Close() error {
	if f.fd != nil {
		if err := f.fd.Close(); err != nil {
			return err
		} else {
			f.fd = nil
		}
	}
	return nil
}

func (f *LazyFile) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	} else if f.fd == nil {
		if err := f.openFile(); err != nil {
			return 0, err
		}
	}
	return f.fd.Read(p)
}

func (f *LazyFile) openFile() error {
	if file, err := os.OpenFile(f.path, os.O_RDWR, 0777); err != nil {
		return err
	} else {
		f.fd = file
		return nil
	}
}
