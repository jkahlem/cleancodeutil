package methodgeneration

import (
	"returntypes-langserver/services/predictor"
	"strings"
)

func CreateMethodDefinition(context predictor.MethodContext, value predictor.MethodValues) string {
	returnType := ConcatTypeName(split(value.ReturnType))
	methodName := ConcatByLowerCamelCase(split(context.MethodName))
	parameterList := ConcatParametersToList(value.Parameters)
	return "public " + returnType + " " + methodName + "(" + parameterList + ") {}"
}

func split(str string) []string {
	return strings.Split(str, " ")
}

func ConcatTypeName(typeName []string) string {
	if len(typeName) == 0 {
		return ""
	}
	if len(typeName) == 1 || typeName[1] == "[]" {
		switch typeName[0] {
		case "void", "int", "float", "double", "byte", "char", "boolean", "long", "short":
			return strings.Join(typeName, "")
		}
	}
	return ConcatByUpperCamelCase(typeName)
}

func ConcatParametersToList(parameters []predictor.Parameter) string {
	output := ""
	for i, par := range parameters {
		if i > 0 {
			output += ", "
		}
		output += ConcatTypeName(split(par.Type))
		if par.IsArray {
			output += "[]"
		}
		output += " " + ConcatByLowerCamelCase(split(par.Name))
	}
	return output
}

func ConcatByLowerCamelCase(words []string) string {
	if len(words) == 1 {
		return words[0]
	}
	return words[0] + ConcatByUpperCamelCase(words[1:])
}

func ConcatByUpperCamelCase(words []string) string {
	output := ""
	for _, word := range words {
		if word == "" {
			continue
		}
		output += strings.ToUpper(word[0:1])
		if len(word) > 1 {
			output += word[1:]
		}
	}
	return output
}
