package methodgeneration

import (
	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/services/predictor"
)

var ErrCouldNotInitialize = errors.ErrorId("Evaluation", "Could not initialize rating method")

type EvaluationSet struct {
	Subsets  []EvaluationSet
	Rater    []Metric
	Filter   configuration.Filter
	Name     string
	Examples []configuration.ExampleGroup
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

func (e *EvaluationSet) initRater(metrics []configuration.MetricConfiguration) errors.Error {
	e.Rater = make([]Metric, 0, len(metrics))
	for _, metric := range metrics {
		switch metric.Type() {
		case configuration.RougeL:
			rater, err := NewRougeLRater(metric)
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, rater)
		case configuration.RougeS:
			rater, err := NewRougeSRater(metric)
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, rater)
		case configuration.RougeN:
			rater, err := NewRougeNRater(metric)
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, rater)
		case configuration.Bleu:
			config, err := metric.AsBleu()
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, &BleuRater{
				config: config,
			})
		case configuration.TokenCounter:
			_, err := metric.AsTokenCounter()
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, &TokenCounter{})
		case configuration.ExactMatch:
			_, err := metric.AsExactMatch()
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, &ExactRater{})
		case configuration.CompilabilityMatch:
			_, err := metric.AsCompilabilityMatch()
			if err != nil {
				return ErrCouldNotInitialize.Wrap(err)
			}
			e.Rater = append(e.Rater, &CompilabilityRater{})
		default:
			return ErrCouldNotInitialize.Wrap(errors.New("Evaluation", "Unknown metric: %s", metric))
		}
	}
	return nil
}

func (e *EvaluationSet) IsMethodAccepted(m Method) bool {
	return csv.IsMethodIncluded(csv.Method{MethodName: m.Name}, e.Filter)
}

func (e *EvaluationSet) GetExampleMethods() ([]predictor.MethodContext, []configuration.MethodExample) {
	methods, examples := mapExampleGroupsToMethod(e.Examples)
	for i := range e.Subsets {
		subsetMethods, subsetExamples := e.Subsets[i].GetExampleMethods()
		if len(subsetMethods) > 0 {
			methods = append(methods, subsetMethods...)
		}
		if len(subsetExamples) > 0 {
			examples = append(examples, subsetExamples...)
		}
	}
	return methods, examples
}
