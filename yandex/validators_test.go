package yandex

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateParsableValue(t *testing.T) {
	correctParseFunc1 := func(value string) error {
		if value != "CORRECT" {
			return fmt.Errorf("expected correct value")
		}
		return nil
	}
	validator1 := validateParsableValue(correctParseFunc1)

	_, es := validator1("CORRECT", "some_key")
	assert.Equal(t, 0, len(es))

	_, es = validator1("INCORRECT", "some_key")
	assert.Equal(t, 1, len(es))

	_, es = validator1([]string{"wrong", "type", "should", "not", "panic"}, "some_key")
	assert.Equal(t, 1, len(es))

	_, es = validator1(666, "some_key")
	assert.Equal(t, 1, len(es))

	correctParseFunc2 := func(value int) (string, error) {
		if value < 500 {
			return "", fmt.Errorf("expected int >= 500")
		}
		return strconv.Itoa(value), nil
	}
	validator2 := validateParsableValue(correctParseFunc2)

	_, es = validator2(777, "some_key")
	assert.Equal(t, 0, len(es))

	_, es = validator2(99, "some_key")
	assert.Equal(t, 1, len(es))

	incorrectParseFunc := func() string {
		return "should not panic"
	}
	validator3 := validateParsableValue(incorrectParseFunc)

	_, es = validator3("something", "some_key")
	assert.Equal(t, 1, len(es))
}

func TestIntGreater(t *testing.T) {
	testCases := []struct {
		val         interface{}
		f           schema.SchemaValidateFunc
		expectedErr *regexp.Regexp
	}{
		{
			val: 1,
			f:   IntGreater(0),
		},
		{
			val:         1.1,
			f:           IntGreater(0),
			expectedErr: regexp.MustCompile("expected type of test_property to be int"),
		},
		{
			val:         0,
			f:           IntGreater(0),
			expectedErr: regexp.MustCompile(`expected test_property to be greater than \(0\), got 0`),
		},
	}

	for i, tc := range testCases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func TestFloatGreater(t *testing.T) {
	testCases := []struct {
		val         interface{}
		f           schema.SchemaValidateFunc
		expectedErr *regexp.Regexp
	}{
		{
			val: float64(1),
			f:   FloatGreater(0),
		},
		{
			val:         int(1),
			f:           FloatGreater(0),
			expectedErr: regexp.MustCompile("expected type of test_property to be float64"),
		},
		{
			val:         float64(0),
			f:           FloatGreater(0),
			expectedErr: regexp.MustCompile(`expected test_property to be greater than \(0.000000\), got 0.000000`),
		},
	}

	for i, tc := range testCases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func TestFloatAtLeast(t *testing.T) {
	testCases := []struct {
		val         interface{}
		f           schema.SchemaValidateFunc
		expectedErr *regexp.Regexp
	}{
		{
			val: float64(0.1),
			f:   FloatAtLeast(0),
		},
		{
			val: float64(0),
			f:   FloatAtLeast(0),
		},
		{
			val:         int(1),
			f:           FloatAtLeast(0),
			expectedErr: regexp.MustCompile("expected type of test_property to be float64"),
		},
		{
			val:         float64(-1),
			f:           FloatAtLeast(0),
			expectedErr: regexp.MustCompile(`expected test_property to be at least \(0.000000\), got -1.000000`),
		},
	}

	for i, tc := range testCases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func matchErr(errs []error, r *regexp.Regexp) bool {
	// err must match one provided
	for _, err := range errs {
		if r.MatchString(err.Error()) {
			return true
		}
	}

	return false
}
