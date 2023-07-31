package yandex

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// generateMapSchemaDiffSuppressFunc Generate MAP DiffSuppressFunc
func generateMapSchemaDiffSuppressFunc(fieldsInfo *objectFieldsInfo) func(k, old, new string, d *schema.ResourceData) bool {
	return func(k, old, new string, d *schema.ResourceData) bool {

		ph := strings.Split(k, ".")

		field := ph[len(ph)-1]

		if fieldsInfo == nil {
			return new == old
		}

		cvf := fieldsInfo.compareValueFunc(field)
		if cvf != nil {
			return cvf(old, new)
		}

		if new == "" {
			return true
		}
		if fieldsInfo.nameFieldsType == nil {
			return new == old
		}

		if tps, ok := fieldsInfo.nameFieldsType[field]; ok {
			for _, t := range tps {
				if checkSuppress(fieldsInfo, field, old, new, t) {
					return true
				}
			}
			return false
		}

		return new == old
	}
}
func checkSuppress(fieldsInfo *objectFieldsInfo, field, old, new string, t reflect.Type) bool {
	fi := fieldsInfo.getType(t, field)
	if fi.valueType == schema.TypeInvalid {
		return false
	}
	if fi.valueType == schema.TypeInt {
		vOld, err := fieldsInfo.stringToInt(field, old)
		if err != nil {
			return false
		}
		vNew, err := fieldsInfo.stringToInt(field, new)
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
	if fi.valueType == schema.TypeFloat {
		vOld, err := fieldsInfo.stringToFloat(field, old)
		if err != nil {
			return false
		}
		vNew, err := fieldsInfo.stringToFloat(field, new)
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
	if fi.valueType == schema.TypeBool {
		vOld, err := fieldsInfo.stringToBool(field, old)
		if err != nil {
			return false
		}
		vNew, err := fieldsInfo.stringToBool(field, new)
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
	if fi.valueType == schema.TypeString {
		return old == new
	}

	return false
}

// generateMapSchemaValidateFunc Generate MAP ValidateFunc
func generateMapSchemaValidateFunc(fieldsInfo *objectFieldsInfo) func(interface{}, string) ([]string, []error) {
	return func(mapRow interface{}, path string) ([]string, []error) {

		fields := make([]string, 0)
		errors := make([]error, 0)

		m := mapRow.(map[string]interface{})

		for k, v := range m {

			cvf := fieldsInfo.checkValueFunc(k)
			if cvf != nil {
				err := cvf(v)
				if err != nil {
					fields = append(fields, k)
					errors = append(errors, fmt.Errorf("Check Fail key %s.%s value: %v err: %v", path, k, v, err))
				}
				continue
			}

			tps, ok := fieldsInfo.nameFieldsType[k]

			if !ok {
				fields = append(fields, k)
				errors = append(errors, fmt.Errorf("Unsupported key %s.%s", path, k))
				continue
			}

			checked := false
			var err error

			for _, t := range tps {
				ok, errOut := checkValidate(fieldsInfo, k, v.(string), t)
				if ok {
					checked = true
					break
				}
				if errOut != nil {
					err = errOut
				}
			}

			if !checked {
				fields = append(fields, k)
				if err == nil {
					errors = append(errors, fmt.Errorf("Check Fail key %s.%s unsupperted type", path, k))
				} else {
					errors = append(errors, fmt.Errorf("Check Fail key %s.%s value: %v err: %v", path, k, v, err))
				}
			}
		}

		return fields, errors
	}
}
func checkValidate(fieldsInfo *objectFieldsInfo, field, value string, t reflect.Type) (bool, error) {
	fi := fieldsInfo.getType(t, field)
	if fi.valueType == schema.TypeInvalid {
		return false, nil
	}
	if fi.valueType == schema.TypeInt {
		i, err := fieldsInfo.stringToInt(field, value)
		if err != nil {
			return false, err
		}
		err = fieldsInfo.intCheckSetValue(field, i)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if fi.valueType == schema.TypeFloat {
		f, err := fieldsInfo.stringToFloat(field, value)
		if err != nil {
			return false, err
		}
		err = fieldsInfo.floatCheckSetValue(field, f)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if fi.valueType == schema.TypeBool {
		b, err := fieldsInfo.stringToBool(field, value)
		if err != nil {
			return false, err
		}
		err = fieldsInfo.boolCheckSetValue(field, b)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	if fi.valueType == schema.TypeString {
		err := fieldsInfo.stringCheckSetValue(field, &value)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, fmt.Errorf("Unsupported field type")
}

func flattenResourceGenerate(fieldsInfo *objectFieldsInfo, v interface{}, includeNil bool,
	useDefault bool, collapseDefault bool,
	knownDefault map[string]struct{}) (map[string]interface{}, error) {

	if v == nil {
		return nil, nil
	}

	t, err := getStructType(v)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}

	res := make(map[string]interface{})
	fields := fieldsInfo.getFields(t)

	for field, fieldInfo := range fields {
		if !fieldsInfo.skip(field) {
			value, err := getValueFrom(v, fieldInfo.name)

			if err != nil {
				return nil, err
			}

			needCollapseDefault := collapseDefault

			if needCollapseDefault && knownDefault != nil {
				if _, ok := knownDefault[field]; ok {
					needCollapseDefault = false
				}
			}
			err = flattenResourceGenerateOneRow(fieldsInfo, fieldInfo, field,
				value, needCollapseDefault, useDefault, includeNil, res)

			if err != nil {
				return nil, err
			}
		}
	}

	return res, nil
}
func flattenResourceGenerateOneRow(fieldsInfo *objectFieldsInfo, fieldInfo fieldReflectInfo, field string,
	value interface{}, collapseDefault bool, useDefault bool, includeNil bool, res map[string]interface{}) error {
	if fieldInfo.valueType == schema.TypeInt {
		if value != nil {
			if collapseDefault && fieldsInfo.intEqualDefault(field, value) {

			} else if fieldsInfo.isStringable(field) {
				vi := value.(int)
				s, err := fieldsInfo.intToString(field, &vi)
				if err != nil {
					return err
				}
				res[field] = s
			} else {
				res[field] = value
			}
		} else if useDefault && fieldsInfo.getDefault(field) != nil {
			res[field] = fieldsInfo.getDefault(field)
		} else if includeNil {
			res[field] = value
		}
	}

	if fieldInfo.valueType == schema.TypeFloat {
		if value != nil {
			res[field] = value
		} else if includeNil {
			res[field] = value
		}
	}

	if fieldInfo.valueType == schema.TypeBool {
		if value != nil {
			res[field] = value
		} else if includeNil {
			res[field] = value
		}
	}

	if fieldInfo.valueType == schema.TypeString {
		if value != nil {
			vs := value.(string)
			if !(!includeNil && collapseDefault && vs == "") {
				res[field] = value
			}
		} else if includeNil {
			res[field] = value
		}
	}

	return nil
}
func flattenResourceGenerateMapS(v interface{}, includeNil bool,
	fieldsInfo *objectFieldsInfo, useDefault bool, collapseDefault bool,
	knownDefault map[string]struct{}) (map[string]string, error) {

	m, err := flattenResourceGenerate(fieldsInfo, v, includeNil, useDefault, collapseDefault, knownDefault)
	if err != nil {
		return nil, err
	}

	if len(m) > 0 {
		out := make(map[string]string)

		for k, v := range m {
			if vI, ok := v.(int); ok {
				out[k] = strconv.Itoa(vI)
			}
			if vF, ok := v.(float64); ok {
				out[k] = strconv.FormatFloat(vF, 'f', -1, 64)
			}
			if vB, ok := v.(bool); ok {
				out[k] = strconv.FormatBool(vB)
			}
			if vS, ok := v.(string); ok {
				out[k] = vS
			}
		}
		return out, nil
	}
	return nil, nil
}

// expandResourceGenerate fill v from resource data
// v must be ptr
func expandResourceGenerateNonSkippedFields(fieldsInfo *objectFieldsInfo, d *schema.ResourceData, v interface{}, path string, skipNil bool) ([]string, error) {

	fieldsOut := []string{}

	if v == nil {
		return fieldsOut, nil
	}

	t, err := getStructType(v)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return fieldsOut, nil
	}

	fields := fieldsInfo.getFields(t)

	for field, fieldInfo := range fields {
		if !fieldsInfo.skip(field) {
			ph := path + field
			// TODO: SA1019: d.GetOkExists is deprecated: usage is discouraged due to undefined behaviors and may be removed in a future version of the SDK (staticcheck)
			value, ok := d.GetOkExists(ph)
			if ok {
				err = expandResourceGenerateOneField(fieldsInfo, fieldInfo, field, v, value, skipNil, ph)
				if err != nil {
					return nil, err
				}
				fieldsOut = append(fieldsOut, field)
			}
		}
	}

	return fieldsOut, nil
}

// expandResourceGenerate fill v from resource data
// v must be ptr
func expandResourceGenerate(fieldsInfo *objectFieldsInfo, d *schema.ResourceData, v interface{}, path string, skipNil bool) error {

	if v == nil {
		return nil
	}

	t, err := getStructType(v)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}

	fields := fieldsInfo.getFields(t)

	for field, fieldInfo := range fields {
		if !fieldsInfo.skip(field) {
			ph := path + field
			// TODO: SA1019: d.GetOkExists is deprecated: usage is discouraged due to undefined behaviors and may be removed in a future version of the SDK (staticcheck)
			value, ok := d.GetOkExists(ph)
			if ok {
				err = expandResourceGenerateOneField(fieldsInfo, fieldInfo, field, v, value, skipNil, ph)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
func expandResourceGenerateOneField(fieldsInfo *objectFieldsInfo, fieldInfo fieldReflectInfo, field string, obj interface{}, value interface{}, skipNil bool, path string) error {
	if value == nil && skipNil {
		return nil
	}

	if fieldInfo.valueType == schema.TypeInt {
		if intValue, ok := value.(int); ok {
			err := fieldsInfo.intCheckSetValue(field, &intValue)
			if err != nil {
				return err
			}

			err = setIntField(obj, fieldInfo.name, &intValue)
			if err != nil {
				return err
			}
		} else if stringValue, ok := value.(string); ok {

			intValue, err := fieldsInfo.stringToInt(field, stringValue)
			if err != nil {
				return err
			}

			if intValue == nil && skipNil {
				return nil
			}

			err = fieldsInfo.intCheckSetValue(field, intValue)
			if err != nil {
				return err
			}

			err = setIntField(obj, fieldInfo.name, intValue)
			if err != nil {
				return err
			}

		} else if value == nil {
			err := fieldsInfo.intCheckSetValue(field, nil)
			if err != nil {
				return err
			}

			err = setIntField(obj, fieldInfo.name, nil)
			if err != nil {
				return err
			}
		} else {
			return &typeMismatchError{text: fmt.Sprintf("expandResourceGenerateOneField: Unknown type for int %s", path)}
		}
	}
	if fieldInfo.valueType == schema.TypeBool {
		if boolValue, ok := value.(bool); ok {
			err := fieldsInfo.boolCheckSetValue(field, &boolValue)
			if err != nil {
				return err
			}

			err = setBoolField(obj, fieldInfo.name, &boolValue)
			if err != nil {
				return err
			}
		} else if stringValue, ok := value.(string); ok {

			boolValue, err := fieldsInfo.stringToBool(field, stringValue)
			if err != nil {
				return err
			}

			if boolValue == nil && skipNil {
				return nil
			}

			err = fieldsInfo.boolCheckSetValue(field, boolValue)
			if err != nil {
				return err
			}

			err = setBoolField(obj, fieldInfo.name, boolValue)
			if err != nil {
				return err
			}

		} else if value == nil {
			err := fieldsInfo.boolCheckSetValue(field, nil)
			if err != nil {
				return err
			}

			err = setBoolField(obj, fieldInfo.name, nil)
			if err != nil {
				return err
			}
		} else {
			return &typeMismatchError{text: fmt.Sprintf("expandResourceGenerateOneField: Unknown type for bool %s", path)}
		}
	}
	if fieldInfo.valueType == schema.TypeFloat {
		if floatValue, ok := value.(float64); ok {
			err := fieldsInfo.floatCheckSetValue(field, &floatValue)
			if err != nil {
				return err
			}

			err = setFloatField(obj, fieldInfo.name, &floatValue)
			if err != nil {
				return err
			}
		} else if stringValue, ok := value.(string); ok {

			floatValue, err := fieldsInfo.stringToFloat(field, stringValue)
			if err != nil {
				return err
			}

			if floatValue == nil && skipNil {
				return nil
			}

			err = fieldsInfo.floatCheckSetValue(field, floatValue)
			if err != nil {
				return err
			}

			err = setFloatField(obj, fieldInfo.name, floatValue)
			if err != nil {
				return err
			}

		} else if value == nil {
			err := fieldsInfo.floatCheckSetValue(field, nil)
			if err != nil {
				return err
			}

			err = setFloatField(obj, fieldInfo.name, nil)
			if err != nil {
				return err
			}
		} else {
			return &typeMismatchError{text: fmt.Sprintf("expandResourceGenerateOneField: Unknown type for float %s", path)}
		}
	}
	if fieldInfo.valueType == schema.TypeString {
		if stringValue, ok := value.(string); ok {

			if stringValue == "" && skipNil {
				return nil
			}

			err := fieldsInfo.stringCheckSetValue(field, &stringValue)
			if err != nil {
				return err
			}

			err = setStringField(obj, fieldInfo.name, &stringValue)
			if err != nil {
				return err
			}

		} else if value == nil {
			return &nilNotAllowedError{text: fmt.Sprintf("expandResourceGenerateOneField: nil  %s", path)}
		} else {
			return &typeMismatchError{text: fmt.Sprintf("expandResourceGenerateOneField: Unknown type for string %s", path)}
		}
	}

	return nil
}
