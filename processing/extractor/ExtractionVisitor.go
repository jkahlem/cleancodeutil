package extractor

import (
	"fmt"
	"returntypes-langserver/common/csv"
	"returntypes-langserver/common/java"
	"returntypes-langserver/common/packagetree"
)

// This visitor visists the elements of a Java Code File and writes each method/class found as a CSV record to the
// methods and classes properties. The type names are resolved to their canonical name (if resolution is possible)
type ExtractionVisitor struct {
	// List of CSV Records for java methods
	methods [][]string
	// List of CSV Records for java classes
	classes [][]string
	// Package tree to use for type resolution
	packageTree *packagetree.Tree
	// The currently visited java code file.
	currentFile *java.CodeFile
}

func (visitor *ExtractionVisitor) VisitCodeFile(codeFile *java.CodeFile) {
	visitor.currentFile = codeFile
	if codeFile.Classes != nil {
		for i := range codeFile.Classes {
			codeFile.Classes[i].Accept(visitor)
		}
	}
}

func (visitor *ExtractionVisitor) VisitClass(class *java.Class) {
	if class.Classes != nil {
		for i := range class.Classes {
			class.Classes[i].Accept(visitor)
		}
	}
	if class.Methods != nil {
		for i := range class.Methods {
			class.Methods[i].Accept(visitor)
		}
	}

	visitor.extractClass(class)
}

// extracts classes with their extended/implemented classes for the classHierarchyWriter using the fully qualified names
func (visitor *ExtractionVisitor) extractClass(class *java.Class) {
	if class.ExtendsImplements != nil {
		visitor.classes = append(visitor.classes, csv.Class{
			ClassName: visitor.currentFile.PackageName + "." + class.ClassName,
			Extends:   visitor.getExtendedClassesWithCanonicalName(class),
		}.ToRecord())
	}
}

// returns a list of the canonical name of each class extended/implemented by the given class
func (visitor *ExtractionVisitor) getExtendedClassesWithCanonicalName(class *java.Class) []string {
	extendedClasses := make([]string, len(class.ExtendsImplements))
	for index := range class.ExtendsImplements {
		extendedClassName := &class.ExtendsImplements[index]
		resolvedClassName, _ := visitor.resolve(extendedClassName)
		extendedClasses[index] = resolvedClassName
	}
	return extendedClasses
}

// resolves the type object to its canonical name. The bool return value is true if the type could be resolved.
func (visitor *ExtractionVisitor) resolve(typeName *java.Type) (string, bool) {
	return java.Resolve(typeName, visitor.packageTree)
}

// Writes methods with their return types into a csvfile
func (visitor *ExtractionVisitor) VisitMethod(method *java.Method) {
	resolvedReturnType, _ := visitor.resolve(&method.ReturnType)
	filePath := ""
	if visitor.currentFile != nil {
		filePath = visitor.currentFile.FilePath
	}
	visitor.methods = append(visitor.methods, csv.Method{
		MethodName: method.MethodName,
		ReturnType: resolvedReturnType,
		Parameters: visitor.extractParameters(method),
		Labels:     java.GetMethodLabels(method),
		FilePath:   filePath,
	}.ToRecord())
}

// Extracts parameters of the method in this format: "<type> <method>"
func (visitor *ExtractionVisitor) extractParameters(method *java.Method) []string {
	result := make([]string, 0, len(method.Parameters))
	for _, parameter := range method.Parameters {
		fmt.Printf(":%s:, :%s:\n", parameter.Type.TypeName, parameter.Name)
		resolvedParameterType, _ := visitor.resolve(&parameter.Type)
		csvStr := fmt.Sprintf("%s %s", resolvedParameterType, parameter.Name)
		result = append(result, csvStr)
	}
	return result
}

func (visitor *ExtractionVisitor) VisitImport(_import *java.Import) {
	// Do nothing.
}

func (visitor *ExtractionVisitor) VisitTypeParameter(typeParameter *java.TypeParameter) {
	// Do nothing.
}
