package base

import "returntypes-langserver/common/debug/errors"

type Creator interface {
	Create() errors.Error
}

type Evaluator interface {
	Evaluate() errors.Error
}

type Trainer interface {
	Train() errors.Error
}

type Dataset struct {
	Path string
}
