package base

import "returntypes-langserver/common/debug/errors"

type Creator interface {
	Create() errors.Error
}
