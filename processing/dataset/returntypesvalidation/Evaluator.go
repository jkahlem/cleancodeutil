package returntypesvalidation

import (
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/processing/dataset/base"
)

type Evaluator struct{}

func NewEvaluator() base.Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(path string) errors.Error {
	return nil
}
