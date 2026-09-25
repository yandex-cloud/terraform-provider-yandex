package value

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

func Any(v Value) (any, error) { //nolint:funlen,gocyclo
	if vv, has := v.(*optionalValue); has {
		if vv == nil || vv.value == nil {
			return nil, nil //nolint:nilnil
		}

		v, err := Any(vv.value)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return v, nil
	}

	switch vv := v.(type) {
	case boolValue:
		return bool(vv), nil
	case uint8Value:
		return uint8(vv), nil
	case int8Value:
		return int8(vv), nil
	case int16Value:
		return int16(vv), nil
	case uint16Value:
		return uint16(vv), nil
	case int32Value:
		return int32(vv), nil
	case *floatValue:
		return vv.value, nil
	case *doubleValue:
		return vv.value, nil
	case bytesValue:
		return []byte(vv), nil
	case *uuidValue:
		// replace to good uuid on migration
		return vv.value, nil
	case uint32Value:
		return uint32(vv), nil
	case dateValue:
		return DateToTime(uint32(vv)), nil
	case datetimeValue:
		return DatetimeToTime(uint32(vv)), nil
	case uint64Value:
		return uint64(vv), nil
	case timestampValue:
		return TimestampToTime(uint64(vv)), nil
	case int64Value:
		return int64(vv), nil
	case intervalValue:
		return IntervalToDuration(int64(vv)), nil
	case tzDateValue:
		t, err := TzDateToTime(string(vv))
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return t, nil
	case tzDatetimeValue:
		t, err := TzDatetimeToTime(string(vv))
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return t, nil
	case tzTimestampValue:
		t, err := TzTimestampToTime(string(vv))
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return t, nil
	case textValue:
		return string(vv), nil
	case dyNumberValue:
		return string(vv), nil
	case ysonValue:
		return []byte(vv), nil
	case jsonValue:
		return xstring.ToBytes(string(vv)), nil
	case jsonDocumentValue:
		return xstring.ToBytes(string(vv)), nil
	default:
		return v, nil
	}
}
