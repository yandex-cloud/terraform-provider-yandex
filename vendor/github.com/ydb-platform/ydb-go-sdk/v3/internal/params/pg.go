package params

import (
	"strconv"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/pg"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
)

type pgParam struct {
	param *Parameter
}

func (p pgParam) Unknown(val string) Builder {
	return p.Value(pg.OIDUnknown, val)
}

func (p pgParam) Value(oid uint32, val string) Builder {
	p.param.value = value.PgValue(oid, val)
	p.param.parent.params = append(p.param.parent.params, p.param)

	return p.param.parent
}

func (p pgParam) Int4(val int32) Builder {
	return p.Value(pg.OIDInt4, strconv.FormatInt(int64(val), 10))
}

func (p pgParam) Int8(val int64) Builder {
	return p.Value(pg.OIDInt8, strconv.FormatInt(val, 10))
}
