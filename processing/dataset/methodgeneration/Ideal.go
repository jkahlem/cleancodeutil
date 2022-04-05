package methodgeneration

import (
	"returntypes-langserver/services/predictor"
	"strings"
)

func CreateMethodDefinition(context predictor.MethodContext, value predictor.MethodValues) string {
	returnType := ConcatByUpperCamelCase(split(value.ReturnType))
	methodName := ConcatByLowerCamelCase(split(context.MethodName))
	parameterList := ConcatParametersToList(value.Parameters)
	return "public " + returnType + " " + methodName + "(" + parameterList + ") {}"
}

func split(str string) []string {
	return strings.Split(str, " ")
}

func ConcatParametersToList(parameters []predictor.Parameter) string {
	output := ""
	for i, par := range parameters {
		if i > 0 {
			output += ", "
		}
		output += ConcatByUpperCamelCase(split(par.Type))
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
		output += strings.ToUpper(word[0:1])
		if len(word) > 1 {
			output += word[1:]
		}
	}
	return output
}
