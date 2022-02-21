package java

import (
	"regexp"
	"returntypes-langserver/common/utils"
	"strings"
)

type MethodLabel string

const (
	Getter           MethodLabel = "getter"
	Setter           MethodLabel = "setter"
	Override         MethodLabel = "override"
	ChainMethod      MethodLabel = "chainMethod"
	ArrayType        MethodLabel = "arrayType"
	TestCode         MethodLabel = "testCode"
	TestMethod       MethodLabel = "testMethod"
	SingleReturn     MethodLabel = "singleReturn"
	SingleAssignment MethodLabel = "singleAssignment"
)

// Creates a list of labels for the given method.
func GetMethodLabels(method *Method) []string {
	labels := make([]string, 0, 3)

	if IsGetter(method) {
		labels = append(labels, string(Getter))
	} else if IsSetter(method) {
		labels = append(labels, string(Setter))
	}
	if IsOverride(method) {
		labels = append(labels, string(Override))
	}
	if method.IsChainMethod {
		labels = append(labels, string(ChainMethod))
	}
	if method.IsSingleReturn {
		labels = append(labels, string(SingleReturn))
	}
	if method.IsSingleAssignment {
		labels = append(labels, string(SingleAssignment))
	}
	if method.ReturnType.IsArrayType {
		labels = append(labels, string(ArrayType))
	}
	if IsInTestFile(method) {
		labels = append(labels, string(TestCode))
	}
	if IsTestMethod(method) {
		labels = append(labels, string(TestMethod))
	}

	return labels
}

// Returns true if the containing file is a test file.
func IsInTestFile(element JavaElement) bool {
	codeFile := FindCodeFile(element)
	if codeFile == nil {
		return false
	}

	matched, _ := regexp.Match(`(T|\bt)est.+\\`, []byte(codeFile.FilePath))
	return matched
}

// Returns true if the method is a test method, so if it has a @Test annotation.
func IsTestMethod(method *Method) bool {
	return utils.ContainsString(method.Annotations, "Test")
}

// Returns true if the method is a getter.
func IsGetter(method *Method) bool {
	if method == nil {
		return false
	}
	return strings.HasPrefix(strings.ToUpper(method.MethodName), "GET")
}

// Returns true if the method is a setter.
func IsSetter(method *Method) bool {
	if method == nil {
		return false
	}
	return strings.HasPrefix(strings.ToUpper(method.MethodName), "SET")
}

// Returns true if the method overrides another method.
func IsOverride(method *Method) bool {
	if method == nil || method.Annotations == nil {
		return false
	}
	for _, annotation := range method.Annotations {
		if annotation == "Override" {
			return true
		}
	}
	return false
}
