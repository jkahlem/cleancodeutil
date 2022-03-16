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

type Evaluator struct {
	Dataset configuration.Dataset
}

type Method struct {
	Name                string
	ExpectedDefinition  string
	GeneratedDefinition string
}

func NewEvaluator(dataset configuration.Dataset) base.Evaluator {
	return &Evaluator{
		Dataset: dataset,
	}
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

	predicted, err := predictor.OnDataset(e.Dataset).GenerateMethods(contexts)
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
		Filter:  setConfiguration.Filter,
	}
	set.initRater(setConfiguration.Metrics)

	for _, subset := range setConfiguration.Subsets {
		set.Subsets = append(set.Subsets, e.buildEvaluationSet(subset))
	}
	return set
}

type EvaluationSet struct {
	Subsets []EvaluationSet
	Rater   []Metric
	Filter  configuration.Filter
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

func (e *EvaluationSet) initRater(metrics []configuration.MetricConfiguration) {
	e.Rater = make([]Metric, 0, len(metrics))
	for _, metric := range metrics {
		switch metric.Type() {
		case configuration.RougeL:
			e.Rater = append(e.Rater, NewRougeLRater(metric))
		case configuration.RougeS:
			e.Rater = append(e.Rater, NewRougeSRater(metric))
		case configuration.RougeN:
			e.Rater = append(e.Rater, NewRougeNRater(metric))
		case configuration.Bleu:
			config, err := metric.AsBleu()
			if err != nil {
				// TODO: remove panic
				panic(err)
			}
			e.Rater = append(e.Rater, &BleuRater{
				config: config,
			})
		default:
			// TODO: remove panic
			panic(fmt.Errorf("Unknown metric: %s", metric))
		}
	}
}

func (e *EvaluationSet) IsMethodAccepted(m Method) bool {
	return csv.IsMethodIncluded(csv.Method{MethodName: m.Name}, e.Filter)
}
