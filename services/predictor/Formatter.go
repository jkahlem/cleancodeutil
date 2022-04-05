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
	}
}

func FormatContexts(contexts []MethodContext, options configuration.SentenceFormattingOptions) {
	if !options.MethodName {
		return
	}
	for i := range contexts {
		if options.MethodName {
			contexts[i].MethodName = string(GetPredictableMethodName(contexts[i].MethodName))
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
