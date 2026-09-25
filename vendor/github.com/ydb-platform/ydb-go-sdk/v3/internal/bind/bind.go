package bind

import (
	"database/sql/driver"
	"sort"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xslices"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type blockID int

const (
	blockDefault = blockID(iota)
	blockPragma
	blockDeclare
	blockYQL
	blockCastArgs
)

type Bind interface {
	ToYdb(sql string, args ...any) (
		yql string, newArgs []any, _ error,
	)

	blockID() blockID
}

type Bindings []Bind

func (bindings Bindings) ToYdb(sql string, args ...any) (
	yql string, pp params.Params, err error,
) {
	if len(bindings) == 0 {
		pp, err = Params(args...)
		if err != nil {
			return "", nil, xerrors.WithStackTrace(err)
		}

		return sql, pp, nil
	}

	if len(args) == 1 {
		if nv, has := args[0].(driver.NamedValue); has {
			if pp, has := nv.Value.(*params.Params); has {
				args = xslices.Transform(*pp, func(v *params.Parameter) any {
					return v
				})
			}
		}
	}

	buffer := xstring.Buffer()
	defer buffer.Free()

	for i := range bindings {
		var err error
		sql, args, err = bindings[len(bindings)-1-i].ToYdb(sql, args...)
		if err != nil {
			return "", nil, xerrors.WithStackTrace(err)
		}
	}

	pp, err = Params(args...)
	if err != nil {
		return "", nil, xerrors.WithStackTrace(err)
	}

	return sql, pp, nil
}

func Sort(bindings []Bind) []Bind {
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].blockID() < bindings[j].blockID()
	})

	return bindings
}
