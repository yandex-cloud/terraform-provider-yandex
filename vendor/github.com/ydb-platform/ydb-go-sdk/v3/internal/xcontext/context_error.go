package xcontext

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
)

var _ error = (*ctxError)(nil)

const (
	atWord   = "at"
	fromWord = "from"
)

func errAt(err error, skipDepth int) error {
	return &ctxError{
		err:         err,
		stackRecord: stack.Record(skipDepth + 1),
		linkWord:    atWord,
	}
}

func errFrom(err error, from string) error {
	return &ctxError{
		err:         err,
		stackRecord: from,
		linkWord:    fromWord,
	}
}

type ctxError struct {
	err         error
	stackRecord string
	linkWord    string
}

func (e *ctxError) Error() string {
	return "'" + e.err.Error() + "' " + e.linkWord + " `" + e.stackRecord + "`"
}

func (e *ctxError) Unwrap() error {
	return e.err
}
