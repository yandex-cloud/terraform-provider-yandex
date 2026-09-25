package budget

import (
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	// ErrNoQuota is a special error for no quota provided by external retry budget
	//
	// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
	ErrNoQuota = xerrors.Wrap(errors.New("no retry quota"))

	errClosedBudget = xerrors.Wrap(errors.New("retry budget closed"))
)
