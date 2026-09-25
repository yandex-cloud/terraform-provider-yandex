package scanner

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type scanStructSettings struct {
	TagName                       string
	AllowMissingColumnsFromSelect bool
	AllowMissingFieldsInStruct    bool
}

type StructScanner struct {
	data *Data
}

func Struct(data *Data) StructScanner {
	return StructScanner{
		data: data,
	}
}

func fieldName(f reflect.StructField, tagName string) string { //nolint:gocritic
	if name, has := f.Tag.Lookup(tagName); has {
		return name
	}

	return f.Name
}

func (s StructScanner) ScanStruct(dst interface{}, opts ...ScanStructOption) (err error) {
	settings := scanStructSettings{
		TagName:                       "sql",
		AllowMissingColumnsFromSelect: false,
		AllowMissingFieldsInStruct:    false,
	}
	for _, opt := range opts {
		if opt != nil {
			opt.applyScanStructOption(&settings)
		}
	}
	ptr := reflect.ValueOf(dst)
	if ptr.Kind() != reflect.Pointer {
		return xerrors.WithStackTrace(fmt.Errorf("%w: '%s'", errDstTypeIsNotAPointer, ptr.Kind().String()))
	}
	if ptr.Elem().Kind() != reflect.Struct {
		return xerrors.WithStackTrace(fmt.Errorf("%w: '%s'", errDstTypeIsNotAPointerToStruct, ptr.Elem().Kind().String()))
	}
	tt := ptr.Elem().Type()
	missingColumns := make([]string, 0, len(s.data.columns))
	existingFields := make(map[string]struct{}, tt.NumField())
	for i := 0; i < tt.NumField(); i++ {
		name := fieldName(tt.Field(i), settings.TagName)
		if name == "-" {
			continue
		}

		v, err := s.data.seekByName(name)
		if err != nil {
			missingColumns = append(missingColumns, name)
		} else {
			if err = value.CastTo(v, ptr.Elem().Field(i).Addr().Interface()); err != nil {
				return xerrors.WithStackTrace(fmt.Errorf("scan error on struct field name '%s': %w", name, err))
			}
			existingFields[name] = struct{}{}
		}
	}

	if !settings.AllowMissingColumnsFromSelect && len(missingColumns) > 0 {
		return xerrors.WithStackTrace(
			fmt.Errorf("%w: '%v'", ErrColumnsNotFoundInRow, strings.Join(missingColumns, "','")),
		)
	}

	if !settings.AllowMissingFieldsInStruct {
		missingFields := make([]string, 0, tt.NumField())
		for _, c := range s.data.columns {
			if _, has := existingFields[c.GetName()]; !has {
				missingFields = append(missingFields, c.GetName())
			}
		}
		if len(missingFields) > 0 {
			return xerrors.WithStackTrace(
				fmt.Errorf("%w: '%v'", ErrFieldsNotFoundInStruct, strings.Join(missingFields, "','")),
			)
		}
	}

	return nil
}
