package ydb

import (
	"database/sql/driver"
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/bind"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xiter"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql"
)

type wrongParameters struct {
	err error
}

func (p wrongParameters) String() string {
	panic(p.err)
}

func (p wrongParameters) Range() xiter.Seq2[string, value.Value] {
	panic(p.err)
}

func (p wrongParameters) ToYDB() (map[string]*Ydb.TypedValue, error) {
	return nil, xerrors.WithStackTrace(p.err)
}

// ParamsFromMap build parameters from named map
func ParamsFromMap(m map[string]any, bindings ...bind.Bind) params.Parameters {
	for _, b := range bindings {
		switch bb := b.(type) {
		case xsql.BindOption:
			switch bb.Bind.(type) {
			case bind.WideTimeTypes:
				continue
			default:
				return wrongParameters{
					err: xerrors.WithStackTrace(fmt.Errorf("%T: %w", b, bind.ErrUnsupportedBindingType)),
				}
			}
		default:
			return wrongParameters{
				err: xerrors.WithStackTrace(fmt.Errorf("%T: %w", b, bind.ErrUnsupportedBindingType)),
			}
		}
	}
	namedParameters := make([]any, 0, len(m))
	for name, val := range m {
		namedParameters = append(namedParameters, driver.NamedValue{Name: name, Value: val})
	}
	_, pp, err := bind.Bindings(bindings).ToYdb("", namedParameters...)
	if err != nil {
		return wrongParameters{err: xerrors.WithStackTrace(err)}
	}

	return &pp
}
