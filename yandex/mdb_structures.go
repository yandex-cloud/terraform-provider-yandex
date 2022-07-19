package yandex

import (
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type MdbConnectionPoolerConfig interface {
	GetPoolingMode() int32
	GetPoolDiscard() *wrapperspb.BoolValue
}

func flattenPGBackupWindowStart(t *timeofday.TimeOfDay) ([]interface{}, error) {
	if t == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["hours"] = int(t.Hours)
	out["minutes"] = int(t.Minutes)

	return []interface{}{out}, nil
}
