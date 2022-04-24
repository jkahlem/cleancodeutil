package methodgeneration

import (
	"fmt"
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
		case configuration.TokenCounter:
			_, err := metric.AsTokenCounter()
			if err != nil {
				// TODO: remove panic
				panic(err)
			}
			e.Rater = append(e.Rater, &TokenCounter{})
		case configuration.ExactMatch:
			_, err := metric.AsExactMatch()
			if err != nil {
				// TODO: remove panic
				panic(err)
			}
			e.Rater = append(e.Rater, &ExactRater{})
		case configuration.CompilabilityMatch:
			_, err := metric.AsCompilabilityMatch()
			if err != nil {
				// TODO: remove panic
				panic(err)
			}
			e.Rater = append(e.Rater, &CompilabilityRater{})
		default:
			// TODO: remove panic
			panic(fmt.Errorf("Unknown metric: %s", metric))
		}
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
