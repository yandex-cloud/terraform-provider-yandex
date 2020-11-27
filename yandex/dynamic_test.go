package yandex

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestDynamic_basic(t *testing.T) {
	t.Parallel()

	us := &postgresql.UserSettings{}

	us.TempFileLimit = &wrappers.Int64Value{Value: 10}

	rv := reflect.ValueOf(us)
	rv = rv.Elem()

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Type().Field(i)

		tg, okTg := FindTag(f, "protobuf", "name")

		if okTg {
			if tg == "default_transaction_isolation" {
				v := 4
				err := setValueToReflect(rv, f.Name, &v)
				if err != nil {
					t.Error(err)
				}
			}
			if tg == "lock_timeout" {
				v := 7
				err := setValueToReflect(rv, f.Name, &v)
				if err != nil {
					t.Error(err)
				}
			}

			if tg == "temp_file_limit" {
				err := setValueToReflect(rv, f.Name, nil)
				if err != nil {
					t.Error(err)
				}
			}

			if tg == "log_statement" {
				err := setValueToReflect(rv, f.Name, nil)
				if err == nil {
					t.Error("setValueToReflect fail: Insert nil into not nil field")
				}
			}
		}

	}

	if us.LockTimeout == nil {
		t.Error("setValueToReflect fail: not set value")
	}

	if us.LockTimeout.GetValue() != 7 {
		t.Error("setValueToReflect fail: value set not correct in *wrappers.Int64Value")
	}

	if us.DefaultTransactionIsolation != 4 {
		t.Error("setValueToReflect fail: not set value in int")
	}

	if us.TempFileLimit != nil {
		t.Error("setValueToReflect fail: not set nil in *wrappers.Int64Value")
	}

	for i := 0; i < rv.NumField(); i++ {
		f := rv.Type().Field(i)

		tg, okTg := FindTag(f, "protobuf", "name")

		if okTg {
			if tg == "default_transaction_isolation" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl.(int) != 4 {
					t.Error("getValueFromReflect fail: read not correct value from int")
				}
			}
			if tg == "lock_timeout" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl.(int) != 7 {
					t.Error("getValueFromReflect fail: read not correct value from *wrappers.Int64Value")
				}
			}

			if tg == "temp_file_limit" {
				vl, err := getValueFromReflect(rv, f.Name)
				if err != nil {
					t.Error(err)
				}
				if vl != nil {
					t.Error("getValueFromReflect read not corect nil value from *wrappers.Int64Value")
				}
			}

		}

	}

}
