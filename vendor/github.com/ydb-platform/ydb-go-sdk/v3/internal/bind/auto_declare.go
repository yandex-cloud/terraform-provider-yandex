package bind

import (
	"sort"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type AutoDeclare struct{}

func (m AutoDeclare) blockID() blockID {
	return blockDeclare
}

func (m AutoDeclare) ToYdb(sql string, args ...any) (
	yql string, newArgs []any, err error,
) {
	params, err := Params(args...)
	if err != nil {
		return "", nil, xerrors.WithStackTrace(err)
	}

	if len(params) == 0 {
		return sql, args, nil
	}

	var (
		declares = make([]string, 0, len(params))
		buffer   = xstring.Buffer()
	)
	defer buffer.Free()

	buffer.WriteString("-- bind declares\n")

	for _, param := range params {
		declares = append(declares, "DECLARE "+param.Name()+" AS "+param.Value().Type().Yql()+";")
	}

	sort.Strings(declares)

	for _, d := range declares {
		buffer.WriteString(d)
		buffer.WriteByte('\n')
	}

	buffer.WriteByte('\n')

	buffer.WriteString(sql)

	for _, param := range params {
		newArgs = append(newArgs, param)
	}

	return buffer.String(), newArgs, nil
}
