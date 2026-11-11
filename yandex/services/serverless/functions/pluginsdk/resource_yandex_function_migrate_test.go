package pluginsdk_test

import (
	"context"
	"reflect"
	"testing"

	svc_cloud_functions "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/pluginsdk"
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

	actual, err := svc_cloud_functions.ResourceYandexFunctionStateUpgradeV0(context.TODO(), partialInputV0, nil)
	if err != nil {
		t.Fatalf("Unexpected error during Function state migration: %s", err)
	}

	if !reflect.DeepEqual(expectedOutputV1, actual) {
		t.Fatalf("Unexpected Function migration result.\nExpected: %#v\nActual: %#v", expectedOutputV1, actual)
	}
}
