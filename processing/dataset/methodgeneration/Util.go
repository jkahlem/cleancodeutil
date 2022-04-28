package methodgeneration

import (
	"returntypes-langserver/common/code/java"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/services/predictor"
	"strings"
)

func mapToMethods(rows []csv.MethodGenerationDatasetRow) ([]predictor.Method, errors.Error) {
	output := make([]predictor.Method, len(rows))
	for i, method := range rows {
		parameters, err := mapToParameters(method.Parameters)
		if err != nil {
			return nil, err
		}
		output[i] = predictor.Method{
			Context: predictor.MethodContext{
				MethodName: method.MethodName,
				ClassName:  strings.Split(method.ClassName, "."),
				IsStatic:   method.IsStatic,
				Types:      method.ContextTypes,
			},
			Values: predictor.MethodValues{
				ReturnType: method.ReturnType,
				Parameters: parameters,
			},
		}
	}
	return output, nil
}

func mapToParameters(rawParameters []string) ([]predictor.Parameter, errors.Error) {
	if csv.IsEmptyList(rawParameters) {
		return nil, nil
	}
	output := make([]predictor.Parameter, len(rawParameters))
	parameters, err := java.ParseParameterList(rawParameters)
	if err != nil {
		return nil, err
	}
	for i, par := range parameters {
		output[i].Type = par.Type.TypeName
		output[i].Name = par.Name
		output[i].IsArray = par.Type.IsArrayType
	}
	return output, nil
}

func mapExampleGroupsToMethod(examples []configuration.ExampleGroup) ([]predictor.MethodContext, []configuration.MethodExample) {
	outputContexts := make([]predictor.MethodContext, 0, len(examples))
	outputExamples := make([]configuration.MethodExample, 0, len(examples))
	for _, exampleGroup := range examples {
		for i := range exampleGroup.Examples {
			if exampleGroup.Examples[i].Label == "" {
				exampleGroup.Examples[i].Label = exampleGroup.Label
			}
		}
		outputContexts = append(outputContexts, mapExamplesToMethod(exampleGroup.Examples)...)
		outputExamples = append(outputExamples, exampleGroup.Examples...)
	}
	return outputContexts, outputExamples
}

func mapExamplesToMethod(examples []configuration.MethodExample) []predictor.MethodContext {
	output := make([]predictor.MethodContext, len(examples))
	for i := range examples {
		output[i] = mapExampleToMethod(examples[i])
	}
	return output
}

func mapExampleToMethod(example configuration.MethodExample) predictor.MethodContext {
	return predictor.MethodContext{
		MethodName: example.MethodName,
		ClassName:  strings.Split(example.ClassName, "."),
		IsStatic:   example.Static,
		Types:      []string{example.ClassName},
	}
}
