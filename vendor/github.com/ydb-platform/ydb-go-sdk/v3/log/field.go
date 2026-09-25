package log

import (
	"fmt"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
)

type (
	Field = kv.KeyValue
)

const (
	IntType      = kv.IntType
	Int64Type    = kv.Int64Type
	StringType   = kv.StringType
	BoolType     = kv.BoolType
	DurationType = kv.DurationType
	StringsType  = kv.StringsType
	ErrorType    = kv.ErrorType
	AnyType      = kv.AnyType
	StringerType = kv.StringerType
)

func appendFieldByCondition(condition bool, ifTrueField Field, fields ...Field) []Field {
	if condition {
		fields = append(fields, ifTrueField)
	}

	return fields
}

func String(k, v string) Field {
	return kv.String(k, v)
}

func Int(k string, v int) Field {
	return kv.Int(k, v)
}

func Int64(k string, v int64) Field {
	return kv.Int64(k, v)
}

func Bool(k string, v bool) Field {
	return kv.Bool(k, v)
}

func Duration(k string, v time.Duration) Field {
	return kv.Duration(k, v)
}

func Strings(k string, v []string) Field {
	return kv.Strings(k, v)
}

func NamedError(k string, v error) Field {
	return kv.NamedError(k, v)
}

func Error(v error) Field {
	return kv.Error(v)
}

func Any(k string, v any) Field {
	return kv.Any(k, v)
}

func Stringer(k string, v fmt.Stringer) Field {
	return kv.Stringer(k, v)
}
