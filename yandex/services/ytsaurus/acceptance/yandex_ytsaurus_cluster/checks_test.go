package yandex_ytsaurus_cluster_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ytsaurus/v1"
	ytsaurusv1sdk "github.com/yandex-cloud/go-sdk/services/ytsaurus/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	ytsaurusClusterResourceName = "yandex_ytsaurus_cluster.test-cluster"
)

func ytsaurusClusterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := ytsaurusv1sdk.NewClusterClient(config.SDKv2).Get(context.Background(), &ytsaurus.GetClusterRequest{
			ClusterId: id,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("project not found")
		}

		return nil
	}
}
