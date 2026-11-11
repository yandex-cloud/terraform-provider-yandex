package sdkcommon

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// parseFunc should take exactly one argument of the type specified in the schema
// and return an error as its last return value
func ValidateParsableValue(parseFunc interface{}) schema.SchemaValidateFunc {
	return func(value interface{}, key string) (warnings []string, errors []error) {
		tryCall := func() (vs []reflect.Value, err error) {
			defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("could not call parse function: %v", p)
				}
			}()

			vs = reflect.ValueOf(parseFunc).Call([]reflect.Value{reflect.ValueOf(value)})
			return
		}

		vs, err := tryCall()
		if err != nil {
			errors = append(errors, err)
			return
		}

		if len(vs) == 0 {
			errors = append(errors, fmt.Errorf("expected parse function to return at least one value"))
			return
		}

		last := vs[len(vs)-1]
		if last.Kind() == reflect.Interface {
			err, ok := last.Interface().(error)
			if ok || last.IsNil() {
				if err != nil {
					errors = append(errors, err)
				}
				return
			}
		}
		errors = append(errors, fmt.Errorf("expected parse function's last return value to be an error"))
		return
	}
}

func ConvertableToInt() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		str, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected %s to be a stringified integer", k))
			return
		}

		_, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			es = append(es, fmt.Errorf("expected %s to be an integer in the range (%d, %d), got %q",
				k, int64(math.MinInt64), int64(math.MaxInt64), str))
			return
		}

		return
	}
}
