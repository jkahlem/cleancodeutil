package utils

import (
	"bufio"
	"fmt"
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
