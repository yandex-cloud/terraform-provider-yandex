package yandex

import (
	"context"
	"reflect"
	"testing"
)

func TestFunctionMigrationRemovesLogGroupId(t *testing.T) {
	t.Parallel()

	partialInputV0 := map[string]any{
		"name":        "function",
		"loggroup_id": "loggroupid",
	}

	expectedOutputV1 := map[string]any{
		"name": "function",
	}

	actual, err := resourceYandexFunctionStateUpgradeV0(context.TODO(), partialInputV0, nil)
	if err != nil {
		t.Fatalf("Unexpected error during Function state migration: %s", err)
	}

	if !reflect.DeepEqual(expectedOutputV1, actual) {
		t.Fatalf("Unexpected Function migration result.\nExpected: %#v\nActual: %#v", expectedOutputV1, actual)
	}
}
