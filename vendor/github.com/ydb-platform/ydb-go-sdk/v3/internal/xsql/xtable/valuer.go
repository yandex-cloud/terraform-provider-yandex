package xtable

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scanner"
)

type valuer struct {
	v interface{}
}

func (v *valuer) UnmarshalYDB(raw scanner.RawValue) error {
	v.v = raw.Any()

	return nil
}

func (v *valuer) Value() interface{} {
	return v.v
}
