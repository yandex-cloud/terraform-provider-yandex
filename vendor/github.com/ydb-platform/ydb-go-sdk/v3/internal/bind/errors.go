package bind

import "errors"

var (
	ErrInconsistentArgs         = errors.New("inconsistent args")
	ErrUnexpectedNumericArgZero = errors.New("unexpected numeric arg $0. Allowed only $1 and greater")
	ErrUnsupportedBindingType   = errors.New("unsupported binding type")
)
