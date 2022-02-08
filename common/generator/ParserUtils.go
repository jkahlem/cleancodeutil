package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

type SourceFilePair struct {
	Source   string
	FileNode *ast.File
}

type context struct {
	files []SourceFilePair
	// The file node which is currently parsed
	currentFile *SourceFilePair
}

// Parses the source code of the passed file
func ParseFile(filePaths ...string) (*context, error) {
	sourceCodes := make([]string, 0, len(filePaths))
	for _, path := range filePaths {
		if src, err := ioutil.ReadFile(path); err != nil {
			return nil, err
		} else {
			sourceCodes = append(sourceCodes, string(src))
		}
	}
	return ParseSourceCode(sourceCodes...)
}

// Parses the passed source code
func ParseSourceCode(sourceCodes ...string) (*context, error) {
	ctx := context{
		files: make([]SourceFilePair, 0, len(sourceCodes)),
	}
	for _, src := range sourceCodes {
		fileNode, err := parser.ParseFile(token.NewFileSet(), "", src, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		ctx.files = append(ctx.files, SourceFilePair{
			Source:   src,
			FileNode: fileNode,
		})
	}
	return &ctx, nil
}

func ParsePackage(directoryPath string) (*context, error) {
	if files, err := os.ReadDir(directoryPath); err != nil {
		return nil, err
	} else {
		paths := make([]string, 0, len(files))
		for _, dir := range files {
			paths = append(paths, filepath.Join(directoryPath, dir.Name()))
		}
		return ParseFile(paths...)
	}
}

// The file from where go generate was called on
func CurrentFile() string {
	return os.Getenv("GOFILE")
}

// Returns the package name for the file. Panics, if no package is specified.
func (ctx *context) Package() string {
	if len(ctx.files) == 0 || ctx.files[0].FileNode == nil || ctx.files[0].FileNode.Name == nil {
		panic("No package specified in the source file.")
	}
	return ctx.files[0].FileNode.Name.Name
}
