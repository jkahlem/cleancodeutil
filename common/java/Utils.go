package java

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/csv"
	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/packagetree"

	"strings"
)

// Maps a slice of names to a slice of types.
func MapStringSliceToTypeSlice(types []string) []Type {
	if len(types) == 1 && types[0] == "" {
		return []Type{}
	}
	result := make([]Type, len(types))
	for i, t := range types {
		result[i].IsArrayType = false
		result[i].TypeName = t
	}
	return result
}

// Adds classes to the package tree as nodes
func FillPackageTreeByCsvClassNodes(dest *packagetree.Tree, classes []csv.Class) {
	if dest == nil {
		return
	}

	for _, class := range classes {
		selector := dest.Select(class.ClassName)
		selector.Parent().Add(CreateNodeForClass(class))
	}
}

// Creates a code file node containing the class.
func CreateNodeForClass(class csv.Class) *CodeFile {
	parts := strings.Split(class.ClassName, ".")
	unqualifiedName := parts[len(parts)-1]
	packageName := strings.Join(parts[:len(parts)-1], ".")

	fileNode := CodeFile{
		PackageName: packageName,
		FilePath:    "<?>",
	}
	classNode := Class{
		ClassName:         unqualifiedName,
		ExtendsImplements: MapStringSliceToTypeSlice(class.Extends),
		Modifiers:         []string{"public"},
		parentElement:     &fileNode,
	}
	fileNode.Classes = []*Class{&classNode}
	return &fileNode
}

// Searches for a file node from the element's parents.
func FindCodeFile(element JavaElement) *CodeFile {
	for element != nil {
		if codeFile, ok := element.(*CodeFile); ok {
			return codeFile
		}
		element = element.Parent()
	}
	return nil
}

// Loads the java standard packages and saves them to the package tree.
func LoadDefaultPackagesToTree(packageTree *packagetree.Tree) errors.Error {
	for _, file := range configuration.DefaultLibraries() {
		records, err := csv.ReadRecords(file)
		if err != nil {
			return err
		}

		classRecords := csv.UnmarshalClasses(records)
		FillPackageTreeByCsvClassNodes(packageTree, classRecords)
	}
	return nil
}

// Loads files of a file container into a package tree.
func LoadFilesToPackageTree(tree *packagetree.Tree, fileContainer FileContainer) errors.Error {
	for _, file := range fileContainer.CodeFiles() {
		if file.PackageName == "" {
			continue
		}
		selector := tree.Select(file.PackageName)
		selector.Add(file)
		if selector.Err() != nil {
			err := errors.New(JavaErrorTitle, "Could not create node in package tree")
			log.ReportProblemWithError(err, "Could not load file %s to package tree", file.FilePath)
		}
	}
	return nil
}
