package methodgeneration

import (
	"fmt"
	"io"
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/services/predictor"
)

type EvaluationSet struct {
	Subsets  []EvaluationSet
	Rater    []Metric
	Filter   configuration.Filter
	Name     string
	Examples []configuration.MethodExample
}

func (e *EvaluationSet) AddMethod(m Method) {
	if !e.IsMethodAccepted(m) {
		return
	}
	for i := range e.Rater {
		e.Rater[i].Rate(m)
	}
	for i := range e.Subsets {
		e.Subsets[i].AddMethod(m)
	}
}

func (e *EvaluationSet) IsIdealScoreRequired() bool {
	for _, r := range e.Rater {
		if r.Name() == "Ideal" { // TODO: Better recognition?
			return true
		}
	}
	for i := range e.Subsets {
		if e.Subsets[i].IsIdealScoreRequired() {
			return true
		}
	}
	return false
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
		case configuration.Ideal:
			_, err := metric.AsIdeal()
			if err != nil {
				// TODO: remove panic
				panic(err)
			}
			e.Rater = append(e.Rater, &IdealRater{})
		default:
			// TODO: remove panic
			panic(fmt.Errorf("Unknown metric: %s", metric))
		}
	}
}

func (e *EvaluationSet) PrintScore(writer io.Writer) {
	if len(e.Rater) > 0 {
		fmt.Fprintf(writer, "Evaluation Type: %s\n", e.Name)
		for i := range e.Rater {
			fmt.Fprintf(writer, "Metric: %s. Score: %v\n", e.Rater[i].Name(), e.Rater[i].Score())
		}
	}
	for i := range e.Subsets {
		e.Subsets[i].PrintScore(writer)
	}
}

func (e *EvaluationSet) IsMethodAccepted(m Method) bool {
	return csv.IsMethodIncluded(csv.Method{MethodName: m.Name}, e.Filter)
}

func (e *EvaluationSet) GetExampleMethods() []predictor.MethodContext {
	methods := mapExamplesToMethod(e.Examples)
	for i := range e.Subsets {
		subsetMethods := e.Subsets[i].GetExampleMethods()
		if len(subsetMethods) > 0 {
			methods = append(methods, subsetMethods...)
		}
	}
	return methods
}
