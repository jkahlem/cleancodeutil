package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
)

type context struct {
	sourceCode string
	fileNode   *ast.File
}

// Parses the source code of the passed file
func ParseFile(filePath string) (*context, error) {
	if src, err := ioutil.ReadFile(filePath); err != nil {
		return nil, err
	} else {
		return ParseSourceCode(string(src))
	}
}

// Parses the passed source code
func ParseSourceCode(src string) (*context, error) {
	srcFileNode, err := parser.ParseFile(token.NewFileSet(), "", src, parser.ParseComments)
	return &context{
		sourceCode: src,
		fileNode:   srcFileNode,
	}, err
}

// The file from where go generate was called on
func CurrentFile() string {
	return os.Getenv("GOFILE")
}

func (ctx *context) Package() string {
	if ctx.fileNode == nil || ctx.fileNode.Name == nil {
		panic("No package specified in the source file.")
	}
	return ctx.fileNode.Name.Name
}
