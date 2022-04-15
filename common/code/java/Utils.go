package java

import (
	"fmt"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"

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
		records, err := csv.NewFileReader(file).ReadClassRecords()
		if err != nil {
			return err
		}

		FillPackageTreeByCsvClassNodes(packageTree, records)
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

const ParameterFieldSeparator = "/"
const ArrayTypeExtension = "[]"

// Formats a slice of parameters into a slice of strings where each string contains the type and name in a specific format.
// The result can be parsed with ParseParameterList back to the parameter slice.
// The nameProvider can be used to do specific formatting on the type name / parameter name before creating the output.
// The nameProvider might be nil - in this case, the type name and parameter name fields are used as they are.
func FormatParameterList(parameters []Parameter, nameProvider func(Parameter) (typ, name string)) []string {
	if nameProvider == nil {
		nameProvider = func(p Parameter) (typ, name string) {
			return p.Type.TypeName, p.Name
		}
	}
	output := make([]string, len(parameters))
	for i, par := range parameters {
		typeName, name := nameProvider(par)
		if par.Type.IsArrayType {
			typeName += ArrayTypeExtension
		}
		output[i] = fmt.Sprintf("%s%s%s", typeName, ParameterFieldSeparator, name)
	}
	return output
}

// Parses a list of parameters formatted with FormatParameterList to the structs.
func ParseParameterList(input []string) ([]Parameter, errors.Error) {
	if len(input) == 1 && input[0] == "" {
		return nil, nil
	}
	output := make([]Parameter, len(input))
	for i, parameter := range input {
		parts := strings.Split(parameter, ParameterFieldSeparator)
		if len(parts) != 2 {
			return nil, errors.New("Parsing Error", "Expected type and name for parameters but got %d value(s).", len(parts))
		}
		output[i] = Parameter{
			Name: parts[1],
			Type: Type{
				TypeName:    strings.TrimRight(parts[0], ArrayTypeExtension),
				IsArrayType: strings.HasSuffix(parts[0], ArrayTypeExtension),
			},
		}
	}
	return output, nil
}
