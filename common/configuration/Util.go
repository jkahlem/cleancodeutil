package configuration

import (
	"os"
	"path/filepath"
)

// Returns the absolute path to the directory of this go project. (derived using the position of the executable)
func GoProjectDir() string {
	dir, err := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), ".."))
	if err != nil {
		return ""
	}
	return dir
}

// Uses the go project directory to fulfill a path to it's absolute path (if it is not already absolute).
// This is needed, as the current working directory may be set to the workspace directory when started by VSC (as extension).
func AbsolutePathFromGoProjectDir(path string) string {
	if filepath.IsAbs(path) || len(path) == 0 {
		return path
	}
	return filepath.Join(GoProjectDir(), path)
}
