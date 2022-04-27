// The predictor package is used for communicating with the predictor application.
// The package defines a high-level API for training the predictor using datasets
// and predict the return types of given method names.
package predictor

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"strings"
)

const PredictorErrorTitle = "Predictor Error"

type SupportedModels string

const (
	ReturnTypesPrediction SupportedModels = "ReturnTypesPrediction"
	MethodGenerator       SupportedModels = "MethodGenerator"
)

type MethodTypeMap map[PredictableMethodName]string

// Interface used for the predictor to support multiple predictor implementations like the mock.
type Predictor interface {
	PredictorGlobal
	// Starts the training and evaluation process.
	TrainReturnTypes(methods []Method, labels [][]string) errors.Error
	// Evaluates the passed methods and returns the scores for it
	EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error)
	// Makes predictions for the methods in the map and sets the types as their value.
	PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error
	// Predicts the expected return type for the given method names. Returns a list of expected return types in the exact order
	// the method names were passed.
	PredictReturnTypes(methodNames []PredictableMethodName) ([]MethodValues, errors.Error)
	// Starts the training and evaluation process. This method might apply side effects on the passed methods.
	TrainMethods(trainingSet []Method) errors.Error
	// Generates the remained part of a method by it's method name. This method might apply side effects on the passed contexts.
	GenerateMethods(contexts []MethodContext) ([][]MethodValues, errors.Error)
	// Returns true if the model exists and is already trained
	ModelExists(modelType SupportedModels) (bool, errors.Error)
	// Returns a list of checkpoints which can be used for OnCheckpoint
	GetCheckpoints(modelType SupportedModels) ([]string, errors.Error)
}

type PredictorGlobal interface {
	GetModels(modelType SupportedModels) ([]Model, errors.Error)
}

type predictor struct {
	config     configuration.Dataset
	checkpoint string
}

func OnDataset(dataset configuration.Dataset) Predictor {
	if configuration.PredictorUseMock() {
		return &mock{}
	}
	return &predictor{
		config: dataset,
	}
}

func OnCheckpoint(dataset configuration.Dataset, checkpoint string) Predictor {
	if configuration.PredictorUseMock() {
		return &mock{}
	}
	return &predictor{
		config:     dataset,
		checkpoint: checkpoint,
	}
}

func (p *predictor) ModelExists(modelType SupportedModels) (bool, errors.Error) {
	options, err := p.getOptions(modelType)
	if err != nil {
		return false, err
	}
	return remote().Exists(options)
}

func (p *predictor) TrainReturnTypes(methods []Method, labels [][]string) errors.Error {
	options, err := p.getOptions(ReturnTypesPrediction)
	if err != nil {
		return err
	}
	options.LabelsCsv = p.asCsvString(labels)
	FormatMethods(methods, p.config.SpecialOptions.SentenceFormatting)
	return remote().Train(methods, options)
}

func (p *predictor) EvaluateReturnTypes(evaluationSet []Method, labels [][]string) (Evaluation, errors.Error) {
	options, err := p.getOptions(ReturnTypesPrediction)
	if err != nil {
		return Evaluation{}, err
	}
	options.LabelsCsv = p.asCsvString(labels)
	FormatMethods(evaluationSet, p.config.SpecialOptions.SentenceFormatting)
	return remote().Evaluate(evaluationSet, options)
}

func (p *predictor) PredictReturnTypes(methodNames []PredictableMethodName) ([]MethodValues, errors.Error) {
	options, err := p.getOptions(ReturnTypesPrediction)
	if err != nil {
		return nil, err
	}

	contexts := make([]MethodContext, len(methodNames))
	for i, name := range methodNames {
		contexts[i].MethodName = string(name)
	}
	return remote().Predict(contexts, options)
}

// Makes predictions for the methods in the map and sets the types as their value.
func (p *predictor) PredictReturnTypesToMap(mapping MethodTypeMap) errors.Error {
	names := p.getMethodNamesInsideOfMap(mapping)
	predictedTypes, err := p.PredictReturnTypes(names)
	if err != nil {
		return err
	}

	if len(names) != len(predictedTypes) {
		return errors.New(PredictorErrorTitle, "Expected %d predictions, but got %d.", len(names), len(predictedTypes))
	}

	for index, name := range names {
		mapping[name] = predictedTypes[index].ReturnType
	}
	return nil
}

func (p *predictor) getMethodNamesInsideOfMap(mapping MethodTypeMap) []PredictableMethodName {
	names := make([]PredictableMethodName, len(mapping))
	i := 0
	for methodName := range mapping {
		names[i] = methodName
		i++
	}
	return names[:i]
}

func (p *predictor) TrainMethods(trainingSet []Method) errors.Error {
	options, err := p.getOptions(MethodGenerator)
	if err != nil {
		return err
	}

	FormatMethods(trainingSet, p.config.SpecialOptions.SentenceFormatting)
	if !p.config.ModelOptions.UseContextTypes {
		for i := range trainingSet {
			trainingSet[i].Context.Types = nil
		}
	}
	return remote().Train(trainingSet, options)
}

func (p *predictor) GenerateMethods(contexts []MethodContext) ([][]MethodValues, errors.Error) {
	options, err := p.getOptions(MethodGenerator)
	if err != nil {
		return nil, err
	}

	FormatContexts(contexts, p.config.SpecialOptions.SentenceFormatting)
	if !p.config.ModelOptions.UseContextTypes {
		for i := range contexts {
			contexts[i].Types = nil
		}
	}
	return remote().PredictMultiple(contexts, options)
}

func (p *predictor) getOptions(modelType SupportedModels) (Options, errors.Error) {
	modelOptions, err := p.mapModelOptions(p.config.ModelOptions)
	if err != nil {
		return Options{}, err
	}
	return Options{
		Identifier:   p.config.QualifiedIdentifier(),
		Type:         modelType,
		ModelOptions: modelOptions,
		Checkpoint:   p.checkpoint,
	}, nil
}

func (p *predictor) mapModelOptions(options configuration.ModelOptions) (ModelOptions, errors.Error) {
	outputOrder, err := p.mapOutputOrder(options.OutputOrder)
	if err != nil {
		return ModelOptions{}, err
	}
	modelOptions := ModelOptions{
		BatchSize:                   options.BatchSize,
		NumOfEpochs:                 options.NumOfEpochs,
		NumReturnSequences:          options.NumReturnSequences,
		MaxSequenceLength:           options.MaxSequenceLength,
		EmptyParameterListByKeyword: options.EmptyParameterListByKeyword,
		Adafactor:                   Adafactor(options.Adafactor),
		Adam:                        Adam(options.Adam),
		ModelType:                   options.ModelType,
		ModelName:                   options.ModelName,
		NumBeams:                    options.NumBeams,
		TopK:                        options.TopK,
		TopN:                        options.TopN,
		LengthPenalty:               options.LengthPenalty,
		OutputOrder:                 outputOrder,
	}
	if options.UseContextTypes {
		modelOptions.DefaultContextTypes = configuration.PredictorDefaultContextTypes()
	}
	return modelOptions, nil
}

const OrderReturnToken = "returnType"
const OrderParameterNameToken = "parameterName"
const OrderParameterTypeToken = "parameterType"

func (p *predictor) mapOutputOrder(order []string) (*OutputComponentOrder, errors.Error) {
	if len(order) == 0 {
		return nil, nil
	}
	returnTypeIndex, err := p.indexOfToken(order, OrderReturnToken)
	if err != nil {
		return nil, err
	}

	parameterNameIndex, err := p.indexOfToken(order, OrderParameterNameToken)
	if err != nil {
		return nil, err
	}

	parameterTypeIndex, err := p.indexOfToken(order, OrderParameterTypeToken)
	if err != nil {
		return nil, err
	}

	if (returnTypeIndex < parameterNameIndex) != (returnTypeIndex < parameterTypeIndex) {
		return nil, errors.New("Error", "Output order pattern: return type token must not come between parameter tokens.")
	}
	return &OutputComponentOrder{
		ReturnType:    returnTypeIndex,
		ParameterName: parameterNameIndex,
		ParameterType: parameterTypeIndex,
	}, nil
}

func (p *predictor) indexOfToken(order []string, token string) (int, errors.Error) {
	for i, t := range order {
		if token == t {
			return i, nil
		}
	}
	return 0, errors.New("Error", "Output order pattern must contain a '%s' token, but it was not found.", token)
}

func (p *predictor) GetCheckpoints(modelType SupportedModels) ([]string, errors.Error) {
	options, err := p.getOptions(MethodGenerator)
	if err != nil {
		return nil, err
	}

	return remote().GetCheckpoints(options)
}

func (p *predictor) GetModels(modelType SupportedModels) ([]Model, errors.Error) {
	return remote().GetModels(modelType)
}

func (p *predictor) asCsvString(records [][]string) string {
	builder := strings.Builder{}
	csv.NewWriter(&builder).WriteAllRecords(records)
	return builder.String()
}
