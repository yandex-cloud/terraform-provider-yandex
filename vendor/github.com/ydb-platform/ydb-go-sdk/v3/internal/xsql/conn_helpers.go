package xsql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Table_V1"

	balancerContext "github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme/helpers"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	internalTable "github.com/ydb-platform/ydb-go-sdk/v3/internal/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xslices"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

func (c *Conn) toYdb(sql string, args ...driver.NamedValue) (yql string, _ *params.Params, _ error) {
	queryArgs := make([]any, len(args))
	for i := range args {
		queryArgs[i] = args[i]
	}

	yql, params, err := c.connector.Bindings().ToYdb(sql, queryArgs...)
	if err != nil {
		return "", nil, xerrors.WithStackTrace(err)
	}

	return yql, &params, nil
}

func (c *Conn) LastUsage() time.Time {
	return c.lastUsage.Get()
}

func (c *Conn) Engine() Engine {
	return c.processor
}

func (c *Conn) GetDatabaseName() string {
	return c.connector.parent.Name()
}

func (c *Conn) normalizePath(tableName string) string {
	return c.connector.pathNormalizer.NormalizePath(tableName)
}

func (c *Conn) tableDescription(ctx context.Context, tableName string) (d options.Description, _ error) {
	d, err := retry.RetryWithResult(ctx,
		func(ctx context.Context) (options.Description, error) {
			d, err := internalTable.DescribeTable(
				balancerContext.WithNodeID(ctx, c.cc.NodeID()),
				c.cc.ID(),
				Ydb_Table_V1.NewTableServiceClient(c.connector.balancer),
				tableName,
			)
			if err != nil {
				return d, xerrors.WithStackTrace(err)
			}

			return d, nil
		},
		retry.WithIdempotent(true),
		retry.WithBudget(c.connector.retryBudget),
		retry.WithTrace(c.connector.traceRetry),
	)
	if err != nil {
		return d, xerrors.WithStackTrace(err)
	}

	return d, nil
}

func (c *Conn) GetColumns(ctx context.Context, tableName string) (columns []string, _ error) {
	d, err := c.tableDescription(ctx, c.normalizePath(tableName))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	for i := range d.Columns {
		columns = append(columns, d.Columns[i].Name)
	}

	return columns, nil
}

func (c *Conn) GetColumnType(ctx context.Context, tableName, columnName string) (dataType string, _ error) {
	d, err := c.tableDescription(ctx, c.normalizePath(tableName))
	if err != nil {
		return "", xerrors.WithStackTrace(err)
	}

	for i := range d.Columns {
		if d.Columns[i].Name == columnName {
			return d.Columns[i].Type.Yql(), nil
		}
	}

	return "", xerrors.WithStackTrace(fmt.Errorf("column '%s' not exist in table '%s'", columnName, tableName))
}

func (c *Conn) GetPrimaryKeys(ctx context.Context, tableName string) ([]string, error) {
	d, err := c.tableDescription(ctx, c.normalizePath(tableName))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return d.PrimaryKey, nil
}

func (c *Conn) IsPrimaryKey(ctx context.Context, tableName, columnName string) (ok bool, _ error) {
	d, err := c.tableDescription(ctx, c.normalizePath(tableName))
	if err != nil {
		return false, xerrors.WithStackTrace(err)
	}

	for _, pkCol := range d.PrimaryKey {
		if pkCol == columnName {
			return true, nil
		}
	}

	return false, nil
}

func (c *Conn) Version(_ context.Context) (_ string, _ error) {
	const version = "default"

	return version, nil
}

func isSysDir(databaseName, dirAbsPath string) bool {
	for _, sysDir := range [...]string{
		path.Join(databaseName, ".sys"),
		path.Join(databaseName, ".sys_health"),
	} {
		if strings.HasPrefix(dirAbsPath, sysDir) {
			return true
		}
	}

	return false
}

func (c *Conn) getTables(ctx context.Context, absPath string, recursive, excludeSysDirs bool) (
	tables []string, _ error,
) {
	if excludeSysDirs && isSysDir(c.connector.parent.Name(), absPath) {
		return nil, nil
	}

	d, err := c.connector.parent.Scheme().ListDirectory(ctx, absPath)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if !d.IsDirectory() && !d.IsDatabase() {
		return nil, xerrors.WithStackTrace(fmt.Errorf("'%s' is not a folder", absPath))
	}

	for i := range d.Children {
		switch d.Children[i].Type {
		case scheme.EntryTable, scheme.EntryColumnTable:
			tables = append(tables, path.Join(absPath, d.Children[i].Name))
		case scheme.EntryDirectory, scheme.EntryDatabase:
			if recursive {
				childTables, err := c.getTables(ctx, path.Join(absPath, d.Children[i].Name), recursive, excludeSysDirs)
				if err != nil {
					return nil, xerrors.WithStackTrace(err)
				}
				tables = append(tables, childTables...)
			}
		}
	}

	return tables, nil
}

func (c *Conn) GetTables(ctx context.Context, folder string, recursive, excludeSysDirs bool) (
	tables []string, _ error,
) {
	absPath := c.normalizePath(folder)

	e, err := c.connector.parent.Scheme().DescribePath(ctx, absPath)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	switch e.Type {
	case scheme.EntryTable, scheme.EntryColumnTable:
		return []string{e.Name}, err
	case scheme.EntryDirectory, scheme.EntryDatabase:
		tables, err = c.getTables(ctx, absPath, recursive, excludeSysDirs)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return xslices.Transform(tables, func(tbl string) string {
			return strings.TrimPrefix(tbl, absPath+"/")
		}), nil
	default:
		return nil, xerrors.WithStackTrace(
			fmt.Errorf("'%s' is not a table or directory (%s)", folder, e.Type.String()),
		)
	}
}

func (c *Conn) GetIndexes(ctx context.Context, tableName string) (indexes []string, _ error) {
	d, err := c.tableDescription(ctx, c.normalizePath(tableName))
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return xslices.Transform(d.Indexes, func(idx options.IndexDescription) string {
		return idx.Name
	}), nil
}

func (c *Conn) GetIndexColumns(ctx context.Context, tableName, indexName string) (columns []string, finalErr error) {
	tableName = c.normalizePath(tableName)
	onDone := trace.DatabaseSQLOnConnGetIndexColumns(c.connector.trace, &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql.(*Conn).GetIndexColumns"),
		tableName, indexName,
	)
	defer func() {
		onDone(columns, finalErr)
	}()
	d, err := c.tableDescription(ctx, tableName)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	for i := range d.Indexes {
		if d.Indexes[i].Name == indexName {
			columns = append(columns, d.Indexes[i].IndexColumns...)
		}
	}

	return xslices.Uniq(columns), nil
}

func (c *Conn) IsColumnExists(ctx context.Context, tableName, columnName string) (columnExists bool, finalErr error) {
	tableName = c.normalizePath(tableName)
	onDone := trace.DatabaseSQLOnConnIsColumnExists(c.connector.trace, &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql.(*Conn).IsColumnExists"),
		tableName, columnName,
	)
	defer func() {
		onDone(columnExists, finalErr)
	}()
	d, err := c.tableDescription(ctx, tableName)
	if err != nil {
		return false, xerrors.WithStackTrace(err)
	}

	for i := range d.Columns {
		if d.Columns[i].Name == columnName {
			return true, nil
		}
	}

	return false, nil
}

func (c *Conn) IsTableExists(ctx context.Context, tableName string) (tableExists bool, finalErr error) {
	tableName = c.normalizePath(tableName)
	onDone := trace.DatabaseSQLOnConnIsTableExists(c.connector.trace, &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql.(*Conn).IsTableExists"),
		tableName,
	)
	defer func() {
		onDone(tableExists, finalErr)
	}()

	tableExists, err := helpers.IsEntryExists(ctx,
		c.connector.parent.Scheme(), tableName,
		scheme.EntryTable, scheme.EntryColumnTable,
	)
	if err != nil {
		return false, xerrors.WithStackTrace(err)
	}

	return tableExists, nil
}
