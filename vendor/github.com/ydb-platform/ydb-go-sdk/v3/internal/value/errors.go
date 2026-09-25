package value

import "errors"

var (
	ErrCannotCast                   = errors.New("cast failed")
	errDestinationTypeIsNotAPointer = errors.New("destination type is not a pointer")
	errNilDestination               = errors.New("destination is nil")
	ErrIssue1501BadUUID             = errors.New("ydb: uuid storage format was broken in go SDK. Now it fixed. And you should select variant for work: typed uuid (good) or use old format with explicit wrapper for read old data") //nolint:lll
)
