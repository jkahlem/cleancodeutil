package predictor

import (
	"fmt"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

type predictorNew struct {
	config configuration.Dataset
}

func OnDataset(dataset configuration.Dataset) *predictorNew {
	return &predictorNew{
		config: dataset,
	}
}

func (p *predictorNew) TrainReturnTypes(methods []Method, labels [][]string) errors.Error {
	options := p.getOptions(ReturnTypesPrediction)
	options.LabelsCsv = p.asCsvString(labels)
	return remote().TrainNew(methods, options)
}

func (p *predictorNew) EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error) {
	options := p.getOptions(ReturnTypesPrediction)
	options.LabelsCsv = p.asCsvString(labels)
	return remote().Evaluate(evaluationSet, options)
}

func (p *predictorNew) PredictReturnType(methodNames []PredictableMethodName) ([]MethodValues, errors.Error) {
	options := p.getOptions(ReturnTypesPrediction)
	contexts := make([]MethodContext, len(methodNames))
	for i, name := range methodNames {
		contexts[i].MethodName = name
	}
	return remote().PredictNew(contexts, options)
}

// Makes predictions for the methods in the map and sets the types as their value.
func (p *predictorNew) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	names := p.getMethodNamesInsideOfMap(mapping)
	predictedTypes, err := p.PredictReturnType(names)
	if err != nil {
		return err
	}

	if len(names) != len(predictedTypes) {
		return errors.New(PredictorErrorTitle, fmt.Sprintf("Expected %d predictions, but got %d.", len(names), len(predictedTypes)))
	}

	for index, name := range names {
		mapping[name] = predictedTypes[index].ReturnType
	}
	return nil
}

func (p *predictorNew) getMethodNamesInsideOfMap(mapping MethodTypeMap) []PredictableMethodName {
	names := make([]PredictableMethodName, len(mapping))
	i := 0
	for methodName := range mapping {
		names[i] = methodName
		i++
	}
	return names[:i]
}

func (p *predictorNew) TrainMethods(trainingSet []Method) errors.Error {
	return remote().TrainNew(trainingSet, p.getOptions(MethodGenerator))
}

func (p *predictorNew) GenerateMethods(contexts []MethodContext) ([]MethodValues, errors.Error) {
	return remote().PredictNew(contexts, p.getOptions(MethodGenerator))
}

func (p *predictorNew) getOptions(modelType SupportedModels) Options {
	return Options{
		Identifier: p.config.QualifiedIdentifier(),
		Type:       modelType,
	}
}

func (p *predictorNew) asCsvString(records [][]string) string {
	builder := strings.Builder{}
	csv.WriteRecordsToTarget(&builder, records)
	return builder.String()
}
