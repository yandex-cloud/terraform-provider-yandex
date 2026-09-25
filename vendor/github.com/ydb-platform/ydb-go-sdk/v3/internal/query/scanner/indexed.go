package scanner

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type IndexedScanner struct {
	data *Data
}

func Indexed(data *Data) IndexedScanner {
	return IndexedScanner{
		data: data,
	}
}

func (s IndexedScanner) Scan(dst ...interface{}) (err error) {
	if len(dst) != len(s.data.columns) {
		return xerrors.WithStackTrace(
			fmt.Errorf("%w: %d != %d",
				errIncompatibleColumnsAndDestinations,
				len(dst), len(s.data.columns),
			),
		)
	}
	for i := range dst {
		v := s.data.seekByIndex(i)
		if err := value.CastTo(v, dst[i]); err != nil {
			return xerrors.WithStackTrace(fmt.Errorf("scan error on column index %d: %w", i, err))
		}
	}

	return nil
}
