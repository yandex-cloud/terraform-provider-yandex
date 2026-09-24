package prompt

import (
	"errors"

	"github.com/cqroot/prompt/constants"
)

var (
	ErrModelConversion = errors.New("model conversion failed")
	ErrUserQuit        = constants.ErrUserQuit
)
