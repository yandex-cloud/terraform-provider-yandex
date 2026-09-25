package bind

import (
	"path"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

type TablePathPrefix string

func (tablePathPrefix TablePathPrefix) blockID() blockID {
	return blockPragma
}

func (tablePathPrefix TablePathPrefix) NormalizePath(folderOrTable string) string {
	switch ch := folderOrTable[0]; ch {
	case '/':
		return folderOrTable
	case '.':
		return path.Join(string(tablePathPrefix), strings.TrimLeft(folderOrTable, "."))
	default:
		return path.Join(string(tablePathPrefix), folderOrTable)
	}
}

func (tablePathPrefix TablePathPrefix) ToYdb(sql string, args ...any) (
	yql string, newArgs []any, err error,
) {
	buffer := xstring.Buffer()
	defer buffer.Free()

	buffer.WriteString("-- bind TablePathPrefix\n")
	buffer.WriteString("PRAGMA TablePathPrefix(\"")
	buffer.WriteString(string(tablePathPrefix))
	buffer.WriteString("\");\n\n")
	buffer.WriteString(sql)

	return buffer.String(), args, nil
}
