package extractor

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/code/packagetree"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/utils"
	"strings"
)

// This visitor visists the elements of a Java Code File and writes each method/class found as a CSV record to the
// methods and classes properties. The type names are resolved to their canonical name (if resolution is possible)
type ExtractionVisitor struct {
	// List of CSV Records for java methods
	methods []csv.Method
	// List of CSV Records for java classes
	classes []csv.Class
	// List of CSV Records for types in files
	fileTypes []csv.FileContextTypes
	// Package tree to use for type resolution
	packageTree *packagetree.Tree
	// The currently visited java code file.
	currentFile *java.CodeFile
	// The currently visited class
	currentClass *java.Class
}

func (visitor *ExtractionVisitor) VisitCodeFile(codeFile *java.CodeFile) {
	visitor.fileTypes = append(visitor.fileTypes, csv.FileContextTypes{
		FilePath:     codeFile.FilePath,
		ContextTypes: make([]string, 0),
	})
	visitor.currentFile = codeFile
	if codeFile.Imports != nil {
		for i := range codeFile.Imports {
			codeFile.Imports[i].Accept(visitor)
		}
	}
	if codeFile.Classes != nil {
		for i := range codeFile.Classes {
			codeFile.Classes[i].Accept(visitor)
		}
	}
}

func (visitor *ExtractionVisitor) VisitClass(class *java.Class) {
	visitor.addContextType(utils.GetStringExtension(class.ClassName, "."))

	if class.Classes != nil {
		for i := range class.Classes {
			class.Classes[i].Accept(visitor)
		}
	}

	visitor.currentClass = class
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
		})
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
		Parameters: visitor.mapParameters(method.Parameters),
		Labels:     java.GetMethodLabels(method),
		FilePath:   filePath,
		ClassName:  visitor.getQualifiedCurrentClassName(),
		Modifier:   method.Modifier,
		ClassField: visitor.findClassFieldMatchInName(method.MethodName),
	})
}

// Searches for the longest match of a class field name in the given name. The match is case insensitive.
// Returns an empty string if no match is found
func (visitor *ExtractionVisitor) findClassFieldMatchInName(name string) string {
	methodNameLower := strings.ToLower(name)
	match := ""
	for _, field := range visitor.currentClass.Fields {
		if len(match) > len(field.Name) {
			continue
		}

		fieldNameLower := strings.ToLower(field.Name)
		if strings.Contains(methodNameLower, fieldNameLower) {
			match = field.Name
		}
	}
	return match
}

// Maps parameters in this format: "<type> <method>"
func (visitor *ExtractionVisitor) mapParameters(parameters []java.Parameter) []string {
	return java.FormatParameterList(parameters, func(p java.Parameter) (typ, name string) {
		if len(p.Type.TypeName) == 1 {
			typeName, isResolved := java.Resolve(&p.Type, visitor.packageTree)
			if isResolved {
				return utils.GetStringExtension(typeName, "."), p.Name
			}
		}
		return p.Type.TypeName, p.Name
	})
}

// Gets the qualified name of the class which is currently visited. The name includes all classes in which it is defined.
// Example:
//   class A {
//	   class B {}
//   }
// The return value for class B would be "A.B"
func (visitor *ExtractionVisitor) getQualifiedCurrentClassName() string {
	if visitor.currentClass == nil {
		return ""
	}
	return visitor.getQualifiedClassName(visitor.currentClass)
}

func (visitor *ExtractionVisitor) getQualifiedClassName(class *java.Class) string {
	if class == nil {
		return ""
	} else if upperClass, ok := class.Parent().(*java.Class); ok && upperClass != nil {
		return visitor.getQualifiedClassName(upperClass) + "." + class.ClassName
	} else {
		return class.ClassName
	}
}

// Gets the unqualified name from an identifier (which might already be unqualified)
func (visitor *ExtractionVisitor) getUnqualifiedTypeName(identifier string) string {
	splitted := strings.Split(identifier, ".")
	return splitted[len(splitted)-1]
}

func (visitor *ExtractionVisitor) VisitImport(_import *java.Import) {
	visitor.addContextType(utils.GetStringExtension(_import.ImportPath, "."))
}

func (visitor *ExtractionVisitor) VisitTypeParameter(typeParameter *java.TypeParameter) {
	// Do nothing.
}

func (visitor *ExtractionVisitor) addContextType(typeName string) {
	l := len(visitor.fileTypes)
	if l > 0 {
		visitor.fileTypes[l-1].ContextTypes = append(visitor.fileTypes[l-1].ContextTypes, typeName)
	}
}
