package utils

import (
	"bufio"
	"fmt"
	"io"
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

// Utility writer which writes output to stdout and to the passed writer
type StdoutPrintWriter struct {
	Writer io.Writer
}

func NewPrintWriter(writer io.Writer) io.Writer {
	return &StdoutPrintWriter{
		Writer: writer,
	}
}

func (w *StdoutPrintWriter) Write(bytes []byte) (n int, err error) {
	fmt.Print(string(bytes))
	return w.Writer.Write(bytes)
}
