package yq_ydb_connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func init() {
	resource.AddTestSweepers("yandex_yq_ydb_connection", &resource.Sweeper{
		Name: "yandex_yq_ydb_connection",
		F:    testSweepYDBConnection,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepYDBConnection(_ string) error {
	return testhelpers.SweepAllConnections(Ydb_FederatedQuery.ConnectionSetting_YDB_DATABASE)
}
