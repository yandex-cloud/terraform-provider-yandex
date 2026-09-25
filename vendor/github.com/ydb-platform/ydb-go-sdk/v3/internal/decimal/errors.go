package decimal

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var errSyntax = xerrors.Wrap(fmt.Errorf("invalid syntax"))

type ParseError struct {
	Err   error
	Input string
}

func (p *ParseError) Error() string {
	return fmt.Sprintf(
		"decimal: parse %q: %v", p.Input, p.Err,
	)
}

func (p *ParseError) Unwrap() error {
	return p.Err
}

func syntaxError(s string) *ParseError {
	return &ParseError{
		Err:   errSyntax,
		Input: s,
	}
}

func precisionError(s string, precision, scale uint32) *ParseError {
	return &ParseError{
		Err:   fmt.Errorf("invalid precision/scale: %d/%d", precision, scale),
		Input: s,
	}
}
