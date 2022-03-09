package base

import (
	"returntypes-langserver/common/dataformat/csv"
	"returntypes-langserver/common/debug/errors"
)

type Evaluator interface {
	Evaluate(path string) errors.Error
}

type Trainer interface {
	Train(path string) errors.Error
}

// Does more specific processings like filters
type MethodProcessor interface {
	Process(*csv.Method) (bool, errors.Error)
	Close() errors.Error
}
