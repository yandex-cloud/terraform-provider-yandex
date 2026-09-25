package scanner

import (
	"fmt"
	"reflect"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type (
	NamedScanner struct {
		data *Data
	}
	namedDestination struct {
		name string
		ref  interface{}
	}
	NamedDestination interface {
		Name() string
		Ref() interface{}
	}
)

func (dst namedDestination) Name() string {
	return dst.name
}

func (dst namedDestination) Ref() interface{} {
	return dst.ref
}

func NamedRef(columnName string, destinationValueReference interface{}) (dst namedDestination) {
	if columnName == "" {
		panic("columnName must be not empty")
	}
	dst.name = columnName
	v := reflect.TypeOf(destinationValueReference)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%T is not reference type", destinationValueReference))
	}
	dst.ref = destinationValueReference

	return dst
}

func Named(data *Data) NamedScanner {
	return NamedScanner{
		data: data,
	}
}

func (s NamedScanner) ScanNamed(dst ...NamedDestination) (err error) {
	for i := range dst {
		v, err := s.data.seekByName(dst[i].Name())
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		if err = value.CastTo(v, dst[i].Ref()); err != nil {
			return xerrors.WithStackTrace(fmt.Errorf("scan error on column name '%s': %w", dst[i].Name(), err))
		}
	}

	return nil
}
