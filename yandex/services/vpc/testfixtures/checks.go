package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CheckNetworkDestroy(config func() (*providerconfig.Config, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cfg, err := config()
		if err != nil {
			return err
		}
		return checkNetworkDestroy(vpcsdk.NewNetworkClient(cfg.SDK), s)
	}
}

type networkGetter interface {
	Get(context.Context, *vpc.GetNetworkRequest, ...grpc.CallOption) (*vpc.Network, error)
}

func checkNetworkDestroy(client networkGetter, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_network" {
			continue
		}
		_, err := client.Get(context.Background(), &vpc.GetNetworkRequest{
			NetworkId: rs.Primary.ID,
		})
		if status.Code(err) == codes.NotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("check yandex_vpc_network %q is deleted: %w", rs.Primary.ID, err)
		}
		return fmt.Errorf("network still exists: %s", rs.Primary.ID)
	}
	return nil
}
