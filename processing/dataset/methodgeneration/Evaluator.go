package methodgeneration

import (
	"fmt"
	"path/filepath"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
	"returntypes-langserver/services/predictor"
)

type Evaluator struct{}

type Method struct {
	Name                string
	ExpectedDefinition  string
	GeneratedDefinition string
}

func NewEvaluator() base.Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	set, err := e.loadEvaluationSet(path)
	if err != nil {
		return err
	}
	methods, err := e.generateMethodDefinitions(set)
	if err != nil {
		return err
	}
	evalset := e.getEvaluationSetConfig()

	for _, m := range methods {
		evalset.AddMethod(m)
	}

	return nil
}

func (e *Evaluator) loadEvaluationSet(path string) ([]predictor.Method, errors.Error) {
	evaluationSet, err := csv.ReadRecords(filepath.Join(path, TrainingSetFileName))
	if err != nil {
		return nil, err
	}
	methods, err := mapToMethods(csv.UnmarshalMethodGenerationDatasetRow(evaluationSet))
	if err != nil {
		return nil, err
	}
	return methods, nil
}

func (e *Evaluator) generateMethodDefinitions(methods []predictor.Method) ([]Method, errors.Error) {
	contexts := make([]predictor.MethodContext, len(methods))
	for i, method := range methods {
		contexts[i] = method.Context
	}

	predicted, err := predictor.OnDataset(configuration.Dataset{}).GenerateMethods(contexts)
	if len(predicted) != len(methods) {
		return nil, errors.New("Predictor error", fmt.Sprintf("Expected %d methods to be generated but got %d.", len(methods), len(predicted)))
	}

	outputMethods := make([]Method, len(predicted))
	for i := range predicted {
		outputMethods[i] = e.parseOutputToMethod(predictor.Method{
			Context: contexts[i],
			Values:  predicted[i],
		}, methods[i].Values)
	}
	return nil, err
}

func (e *Evaluator) parseOutputToMethod(method predictor.Method, expectedValues predictor.MethodValues) Method {
	return Method{}
}

func (e *Evaluator) getEvaluationSetConfig() *EvaluationSet {
	set := e.buildEvaluationSet(configuration.EvaluationSet{
		Subsets: configuration.EvaluationSubsets(),
	})
	return &set
}

func (e *Evaluator) buildEvaluationSet(setConfiguration configuration.EvaluationSet) EvaluationSet {
	set := EvaluationSet{
		Subsets: make([]EvaluationSet, 0),
	}
	set.initRater(setConfiguration.RatingTypes)

	for _, subset := range setConfiguration.Subsets {
		set.Subsets = append(set.Subsets, e.buildEvaluationSet(subset))
	}
	return set
}

func (e *Evaluator) getAvailableRater() []Rater {
	// TODO: Use configuration.EvaluationRatingTypes() to determine which rater to add
	return []Rater{&AllZeroRater{}}
}

type EvaluationSet struct {
	Subsets []EvaluationSet
	Rater   []Rater
}

func (e *EvaluationSet) AddMethod(m Method) {
	if !e.IsMethodAccepted(m) {
		return
	}
	// TOOD:
	// - Add to output?
	for i := range e.Rater {
		e.Rater[i].Rate(m)
	}
	for i := range e.Subsets {
		e.Subsets[i].AddMethod(m)
	}
}

func (e *EvaluationSet) initRater(ratingTypes []string) {
	e.Rater = make([]Rater, 0, len(ratingTypes))
	for _, ratingType := range ratingTypes {
		switch ratingType {
		case "rouge-l":
			e.Rater = append(e.Rater, &RougeRater{Type: RougeL})
		case "rouge-s":
			e.Rater = append(e.Rater, &RougeRater{Type: RougeS})
		case "rouge-n":
			e.Rater = append(e.Rater, &RougeRater{Type: RougeN})
		case "bleu":
			e.Rater = append(e.Rater, &BleuRater{})
		default:
			// TODO: remove panic
			panic(fmt.Errorf("Unknown rating type: %s", ratingType))
		}
	}
}

func (e *EvaluationSet) IsMethodAccepted(m Method) bool {
	// TODO
	return true
}
