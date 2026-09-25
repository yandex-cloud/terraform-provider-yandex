package query

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xiter"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

var (
	_ query.ResultSet = (*resultSet)(nil)
	_ query.ResultSet = (*materializedResultSet)(nil)
)

type (
	materializedResultSet struct {
		index       int
		columnNames []string
		columnTypes []types.Type
		rows        []query.Row
		rowIndex    int
	}
	resultSet struct {
		index               int64
		recv                func() (*Ydb_Query.ExecuteQueryResponsePart, error)
		columns             []*Ydb.Column
		currentPart         *Ydb_Query.ExecuteQueryResponsePart
		rowIndex            int
		done                chan struct{}
		mustBeLastResultSet bool
	}
	resultSetWithClose struct {
		*resultSet
		close func(ctx context.Context) error
	}
)

func rangeRows(ctx context.Context, rs result.Set) xiter.Seq2[result.Row, error] {
	return func(yield func(result.Row, error) bool) {
		for {
			rs, err := rs.NextRow(ctx)
			if err != nil {
				if xerrors.Is(err, io.EOF) {
					return
				}
			}
			cont := yield(rs, err)
			if !cont || err != nil {
				return
			}
		}
	}
}

func (*materializedResultSet) Close(context.Context) error {
	return nil
}

func (rs *resultSetWithClose) Close(ctx context.Context) error {
	return rs.close(ctx)
}

func (rs *materializedResultSet) Rows(ctx context.Context) xiter.Seq2[result.Row, error] {
	return rangeRows(ctx, rs)
}

func (rs *resultSet) Rows(ctx context.Context) xiter.Seq2[result.Row, error] {
	return rangeRows(ctx, rs)
}

func (rs *materializedResultSet) Columns() (columnNames []string) {
	return rs.columnNames
}

func (rs *materializedResultSet) ColumnTypes() []types.Type {
	return rs.columnTypes
}

func (rs *resultSet) ColumnTypes() (columnTypes []types.Type) {
	columnTypes = make([]types.Type, len(rs.columns))
	for i := range rs.columns {
		columnTypes[i] = types.TypeFromYDB(rs.columns[i].GetType())
	}

	return columnTypes
}

func (rs *resultSet) Columns() (columnNames []string) {
	columnNames = make([]string, len(rs.columns))
	for i := range rs.columns {
		columnNames[i] = rs.columns[i].GetName()
	}

	return columnNames
}

func (rs *materializedResultSet) NextRow(ctx context.Context) (query.Row, error) {
	if rs.rowIndex == len(rs.rows) {
		return nil, xerrors.WithStackTrace(io.EOF)
	}

	defer func() {
		rs.rowIndex++
	}()

	return rs.rows[rs.rowIndex], nil
}

func (rs *materializedResultSet) Index() int {
	if rs == nil {
		return -1
	}

	return rs.index
}

func MaterializedResultSet(
	index int,
	columnNames []string,
	columnTypes []types.Type,
	rows []query.Row,
) *materializedResultSet {
	return &materializedResultSet{
		index:       index,
		columnNames: columnNames,
		columnTypes: columnTypes,
		rows:        rows,
	}
}

func newResultSet(
	recv func() (*Ydb_Query.ExecuteQueryResponsePart, error),
	part *Ydb_Query.ExecuteQueryResponsePart,
) *resultSet {
	return &resultSet{
		index:       part.GetResultSetIndex(),
		recv:        recv,
		currentPart: part,
		rowIndex:    -1,
		columns:     part.GetResultSet().GetColumns(),
		done:        make(chan struct{}),
	}
}

func (rs *resultSet) nextRow(ctx context.Context) (*Row, error) {
	rs.rowIndex++
	for {
		select {
		case <-rs.done:
			return nil, io.EOF
		case <-ctx.Done():
			return nil, xerrors.WithStackTrace(ctx.Err())
		default:
			//nolint:nestif
			if rs.rowIndex == len(rs.currentPart.GetResultSet().GetRows()) {
				part, err := rs.recv()
				if err != nil {
					if xerrors.Is(err, io.EOF) {
						close(rs.done)
					}

					if rs.mustBeLastResultSet && errors.Is(err, errReadNextResultSet) {
						// prevent detect io.EOF in the error
						return nil, xerrors.WithStackTrace(xerrors.Wrap(errors.New(err.Error())))
					}

					return nil, xerrors.WithStackTrace(err)
				}
				rs.rowIndex = 0
				rs.currentPart = part
				if part == nil {
					close(rs.done)

					return nil, xerrors.WithStackTrace(io.EOF)
				}
			}
			if rs.currentPart.GetResultSet() != nil && rs.index != rs.currentPart.GetResultSetIndex() {
				close(rs.done)

				return nil, xerrors.WithStackTrace(fmt.Errorf(
					"received part with result set index = %d, current result set index = %d: %w",
					rs.index, rs.currentPart.GetResultSetIndex(), errWrongResultSetIndex,
				))
			}

			if rs.rowIndex < len(rs.currentPart.GetResultSet().GetRows()) {
				return NewRow(rs.columns, rs.currentPart.GetResultSet().GetRows()[rs.rowIndex]), nil
			}
		}
	}
}

func (rs *resultSet) NextRow(ctx context.Context) (_ query.Row, err error) {
	return rs.nextRow(ctx)
}

func (rs *resultSet) Index() int {
	if rs == nil {
		return -1
	}

	return int(rs.index)
}
