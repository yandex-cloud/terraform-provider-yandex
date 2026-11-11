package sdkcommon_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func TestValidateParsableValue(t *testing.T) {
	correctParseFunc1 := func(value string) error {
		if value != "CORRECT" {
			return fmt.Errorf("expected correct value")
		}
		return nil
	}
	validator1 := sdkcommon.ValidateParsableValue(correctParseFunc1)

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
	validator2 := sdkcommon.ValidateParsableValue(correctParseFunc2)

	_, es = validator2(777, "some_key")
	assert.Equal(t, 0, len(es))

	_, es = validator2(99, "some_key")
	assert.Equal(t, 1, len(es))

	incorrectParseFunc := func() string {
		return "should not panic"
	}
	validator3 := sdkcommon.ValidateParsableValue(incorrectParseFunc)

	_, es = validator3("something", "some_key")
	assert.Equal(t, 1, len(es))
}
