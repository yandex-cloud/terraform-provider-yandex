package pluginsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IntGreater returns a SchemaValidateFunc which tests if the provided value
// is of type int and is greater than provided min (not inclusive)
func IntGreater(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(int)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be int", k))
			return nil, errors
		}

		if v <= min {
			errors = append(errors, fmt.Errorf("expected %s to be greater than (%d), got %d", k, min, v))
			return nil, errors
		}

		return nil, errors
	}
}

// FloatAtLeast returns a SchemaValidateFunc which tests if the provided value
// is of type float64 and is greater than provided min (not inclusive)
func FloatGreater(min float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(float64)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be float64", k))
			return nil, errors
		}

		if v <= min {
			errors = append(errors, fmt.Errorf("expected %s to be greater than (%f), got %f", k, min, v))
			return nil, errors
		}

		return nil, errors
	}
}
