package yq_yds_connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func init() {
	resource.AddTestSweepers("yandex_yq_yds_connection", &resource.Sweeper{
		Name: "yandex_yq_yds_connection",
		F:    testSweepYDSConnection,
		Dependencies: []string{
			"yandex_yq_yds_binding",
		},
	})

	resource.AddTestSweepers("yandex_yq_yds_binding", &resource.Sweeper{
		Name: "yandex_yq_yds_binding",
		F:    testSweepYDSBinding,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepYDSConnection(_ string) error {
	return testhelpers.SweepAllConnections(Ydb_FederatedQuery.ConnectionSetting_DATA_STREAMS)
}

func testSweepYDSBinding(_ string) error {
	return testhelpers.SweepAllBindings(Ydb_FederatedQuery.BindingSetting_DATA_STREAMS)
}
