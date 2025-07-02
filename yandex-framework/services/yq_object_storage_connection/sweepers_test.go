package yq_object_storage_connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func init() {
	resource.AddTestSweepers("yandex_yq_object_storage_connection", &resource.Sweeper{
		Name: "yandex_yq_object_storage_connection",
		F:    testSweepObjectStorageConnection,
		Dependencies: []string{
			"yandex_yq_object_storage_binding",
		},
	})

	resource.AddTestSweepers("yandex_yq_object_storage_binding", &resource.Sweeper{
		Name: "yandex_yq_object_storage_binding",
		F:    testSweepObjectStorageBinding,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepObjectStorageConnection(_ string) error {
	return testhelpers.SweepAllConnections(Ydb_FederatedQuery.ConnectionSetting_OBJECT_STORAGE)
}

func testSweepObjectStorageBinding(_ string) error {
	return testhelpers.SweepAllBindings(Ydb_FederatedQuery.BindingSetting_OBJECT_STORAGE)
}
