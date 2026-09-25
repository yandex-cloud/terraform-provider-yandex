package scanner

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type Data struct {
	columns []*Ydb.Column
	values  []*Ydb.Value
}

func NewData(columns []*Ydb.Column, values []*Ydb.Value) *Data {
	return &Data{
		columns: columns,
		values:  values,
	}
}

func (d Data) seekByName(name string) (value.Value, error) {
	for i := range d.columns {
		if d.columns[i].GetName() == name {
			return value.FromYDB(d.columns[i].GetType(), d.values[i]), nil
		}
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf("'%s': %w", name, ErrColumnsNotFoundInRow))
}

func (d Data) seekByIndex(idx int) value.Value {
	return value.FromYDB(d.columns[idx].GetType(), d.values[idx])
}

func (d Data) Values() []value.Value {
	values := make([]value.Value, len(d.columns))

	for idx := range values {
		values[idx] = value.FromYDB(d.columns[idx].GetType(), d.values[idx])
	}

	return values
}
