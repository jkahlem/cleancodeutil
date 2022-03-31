package methodgeneration

import (
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
				MethodName: predictor.PredictableMethodName(method.MethodName),
				ClassName:  method.ClassName,
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

func mapToParameters(parameters []string) ([]predictor.Parameter, errors.Error) {
	if parameters[0] == "void" {
		return nil, nil
	}
	output := make([]predictor.Parameter, len(parameters))
	for i := range parameters {
		typeAndNamePair := strings.Split(parameters[i], " ")
		if len(typeAndNamePair) != 2 {
			return nil, errors.New("Format error", "Unexpected format for parameters in dataset output: %s", parameters[i])
		}
		output[i].Type = typeAndNamePair[0]
		output[i].Name = typeAndNamePair[1]
	}
	return output, nil
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
		MethodName: predictor.PredictableMethodName(predictor.SplitMethodNameToSentence(example.MethodName)),
		ClassName:  example.ClassName,
		IsStatic:   example.Static,
		Types:      []string{example.ClassName},
	}
}
