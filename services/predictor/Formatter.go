package predictor

import "returntypes-langserver/common/configuration"

func FormatMethods(methods []Method, options configuration.SentenceFormattingOptions) {
	if !options.MethodName && !options.ParameterName && !options.TypeName {
		return
	}
	for i := range methods {
		if options.MethodName {
			methods[i].Context.MethodName = string(GetPredictableMethodName(methods[i].Context.MethodName))
		}
		if options.TypeName {
			methods[i].Values.ReturnType = string(GetPredictableMethodName(methods[i].Values.ReturnType))
		}
		FormatParameters(methods[i].Values.Parameters, options)
		FormatClassNames(methods[i].Context.ClassName, options)
	}
}

func FormatContexts(contexts []MethodContext, options configuration.SentenceFormattingOptions) {
	if !options.MethodName && !options.TypeName {
		return
	}
	for i := range contexts {
		if options.MethodName {
			contexts[i].MethodName = string(GetPredictableMethodName(contexts[i].MethodName))
		}
		FormatClassNames(contexts[i].ClassName, options)
	}
}

func FormatClassNames(classNames []string, options configuration.SentenceFormattingOptions) {
	if !options.TypeName {
		return
	}
	for i, name := range classNames {
		classNames[i] = string(GetPredictableMethodName(name))
	}
}

func FormatValues(values [][]MethodValues, options configuration.SentenceFormattingOptions) {
	if !options.TypeName {
		return
	}
	for i := range values {
		for j := range values[i] {
			if options.TypeName {
				values[i][j].ReturnType = string(GetPredictableMethodName(values[i][j].ReturnType))
			}
			FormatParameters(values[i][j].Parameters, options)
		}
	}
}

func FormatParameters(parameters []Parameter, options configuration.SentenceFormattingOptions) {
	for i := range parameters {
		if options.ParameterName {
			parameters[i].Name = string(GetPredictableMethodName(parameters[i].Name))
		}
		if options.TypeName {
			parameters[i].Type = string(GetPredictableMethodName(parameters[i].Type))
		}
	}
}
