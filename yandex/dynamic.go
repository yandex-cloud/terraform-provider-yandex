package yandex

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// FindTag find tag in field
func FindTag(f reflect.StructField, tagName, name string) (string, bool) {
	tagF := f.Tag.Get(tagName)
	if tagF == "" {
		return "", false
	}

	tags := strings.Split(tagF, ",")
	for _, tag := range tags {
		kvs := strings.SplitN(tag, "=", 2)
		if kvs[0] != name {
			continue
		}
		val := ""
		if len(kvs) == 2 {
			val = kvs[1]
		}
		return val, true
	}
	return "", false
}

func setValueToReflect(rv reflect.Value, name string, v *int) error {

	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		if v == nil {
			return fmt.Errorf("field %s is not nilable", name)
		}
		f.SetInt(int64(*v))
		return nil
	case reflect.Ptr:
		if f.Type() == wrapperspbInt64Value() {
			if v == nil {
				var w *wrappers.Int64Value
				f.Set(reflect.ValueOf(w))
				return nil
			}

			w := wrappers.Int64Value{
				Value: int64(*v),
			}
			f.Set(reflect.ValueOf(&w))
			return nil

		}
		return fmt.Errorf("setValueInt type ptr not implement")
	default:
		return fmt.Errorf("setValueInt type not implement")
	}
}

func getValueFromReflect(rv reflect.Value, name string) (interface{}, error) {
	f := rv.FieldByName(name)

	switch f.Type().Kind() {
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		return int(f.Int()), nil
	case reflect.Ptr:
		if f.Type() == wrapperspbInt64Value() {
			if f.IsNil() {
				return nil, nil
			}
			v, ok := f.Interface().(*wrappers.Int64Value)
			if ok {
				return int(v.GetValue()), nil
			}
			return nil, fmt.Errorf("getValueFromReflect: type mismatch")
		}
		return nil, fmt.Errorf("getValueFromReflect: type ptr not implement")
	default:
		return nil, fmt.Errorf("getValueFromReflect type not implement")
	}
}

func generateMapSchemaDiffSuppressFunc(fai *objectFieldsAdditionalInfo) func(k, old, new string, d *schema.ResourceData) bool {
	return func(k, old, new string, d *schema.ResourceData) bool {
		if new == "" {
			return true
		}

		ph := strings.Split(k, ".")

		field := ph[len(ph)-1]

		if fai.getType(field) == schema.TypeInt {
			vOld, err := fai.iToInt(field, old)
			if err != nil {
				return false
			}
			vNew, err := fai.iToInt(field, new)
			if err != nil {
				return false
			}
			if vOld == nil && vNew == nil {
				return true
			}
			if vNew == nil {
				return true
			}

			return vOld != nil && vNew != nil && *vOld == *vNew
		}

		return new == old
	}
}

func generateMapSchemaValidateFunc(fai *objectFieldsAdditionalInfo) func(interface{}, string) ([]string, []error) {
	return func(mapRow interface{}, path string) ([]string, []error) {

		f := make([]string, 0)
		e := make([]error, 0)

		m := mapRow.(map[string]interface{})

		for k, v := range m {
			if fai.getType(k) == schema.TypeInt {
				s := v.(string)
				i, err := fai.iToInt(k, s)
				if err != nil {
					f = append(f, k)
					e = append(e, fmt.Errorf("Invalid value in %s.%s err: %v", path, k, err))
					continue
				}
				err = fai.iCheckSetValue(k, i)
				if err != nil {
					f = append(f, k)
					e = append(e, fmt.Errorf("Invalid check value in %s.%s err: %v", path, k, err))
					continue
				}
			} else {
				f = append(f, k)
				e = append(e, fmt.Errorf("Unsupperted key %s.%s", path, k))
			}
		}

		return f, e
	}
}

func wrapperspbInt64Value() reflect.Type {
	return reflect.TypeOf(&wrappers.Int64Value{})
}

func flattenResourceGenerateMapS(v interface{}, includeNil bool,
	fai *objectFieldsAdditionalInfo, useDefault bool, collapseDefault bool) (map[string]string, error) {

	m, err := flattenResourceGenerate(v, includeNil, fai, useDefault, collapseDefault)
	if err != nil {
		return nil, err
	}

	if len(m) > 0 {
		out := make(map[string]string)

		for k, v := range m {
			if vI, ok := v.(int); ok {
				out[k] = strconv.Itoa(vI)
			}
			if vS, ok := v.(string); ok {
				out[k] = vS
			}
		}
		return out, nil
	}
	return nil, nil
}

func flattenResourceGenerate(v interface{}, includeNil bool,
	fai *objectFieldsAdditionalInfo, useDefault bool, collapseDefault bool) (map[string]interface{}, error) {

	rv := reflect.ValueOf(v)

	return flattenResourceGenerateReflect(rv, includeNil, fai, useDefault, collapseDefault)
}

func flattenResourceGenerateReflect(rv reflect.Value, includeNil bool,
	fai *objectFieldsAdditionalInfo, useDefault bool, collapseDefault bool) (map[string]interface{}, error) {

	t := rv.Type()

	if t.Kind() == reflect.Ptr && rv.IsNil() {
		return nil, nil
	}
	if t.Kind() == reflect.Ptr {
		return flattenResourceGenerateReflect(rv.Elem(), includeNil, fai, useDefault, collapseDefault)
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("flattenResourceGenerateReflect: not implement type")
	}

	res := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tg, okTg := FindTag(f, "protobuf", "name")

		if okTg {
			if !fai.skip(tg) {
				if f.Type.Kind() == reflect.Int32 || f.Type == wrapperspbInt64Value() {
					itm, err := getValueFromReflect(rv, f.Name)
					if err != nil {
						return nil, err
					}
					if itm != nil {
						if collapseDefault && fai.iEqualDefault(tg, itm) {

						} else if fai.isIStringable(tg) {
							vi := itm.(int)
							s, err := fai.iToString(tg, &vi)
							if err != nil {
								return nil, err
							}
							res[tg] = s
						} else {
							res[tg] = itm
						}
					} else if useDefault && fai.iGetDefault(tg) != nil {
						res[tg] = fai.iGetDefault(tg)
					} else if includeNil {
						res[tg] = itm
					}

				}
			}
		}
	}

	return res, nil
}

func expandResourceGenerateDReflect(rv reflect.Value, d *schema.ResourceData, path string, fai *objectFieldsAdditionalInfo, skipNil bool) error {
	if rv.Kind() == reflect.Ptr {
		return expandResourceGenerateDReflect(rv.Elem(), d, path, fai, skipNil)
	}
	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tg, okTg := FindTag(f, "protobuf", "name")

		if okTg {
			if f.Type.Kind() == reflect.Int32 || f.Type == wrapperspbInt64Value() {
				ph := path + tg
				vSrc, ok := d.GetOkExists(ph)

				if ok {
					if vSrcInt, ok := vSrc.(int); ok {
						err := fai.iCheckSetValue(tg, &vSrcInt)
						if err != nil {
							return err
						}

						err = setValueToReflect(rv, f.Name, &vSrcInt)
						if err != nil {
							return err
						}
					} else if vSrcStr, ok := vSrc.(string); ok {

						vSrcInt, err := fai.iToInt(tg, vSrcStr)
						if err != nil {
							return err
						}

						if vSrcInt == nil && skipNil {
							continue
						}

						err = fai.iCheckSetValue(tg, vSrcInt)
						if err != nil {
							return err
						}

						if vSrcInt == nil && f.Type.Kind() == reflect.Int32 {
							return fmt.Errorf("expandResourceGenerateDReflect: value can't be nil for %s", ph)
						}

						err = setValueToReflect(rv, f.Name, vSrcInt)
						if err != nil {
							return err
						}

					} else {
						return fmt.Errorf("expandResourceGenerateDReflect: Unknown type for %s", ph)
					}
				}
			}
		}
	}

	return nil
}

func expandResourceGenerateD(d *schema.ResourceData, v interface{}, path string, fai *objectFieldsAdditionalInfo, skipNil bool) error {

	if v == nil {
		return nil
	}

	return expandResourceGenerateDReflect(reflect.ValueOf(v), d, path, fai, skipNil)

}
