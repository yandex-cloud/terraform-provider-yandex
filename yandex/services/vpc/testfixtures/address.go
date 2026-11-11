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

func CheckAddressDestroy(config func() (*providerconfig.Config, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cfg, err := config()
		if err != nil {
			return err
		}
		return checkAddressDestroy(vpcsdk.NewAddressClient(cfg.SDK), s)
	}
}

type addressGetter interface {
	Get(context.Context, *vpc.GetAddressRequest, ...grpc.CallOption) (*vpc.Address, error)
}

func checkAddressDestroy(client addressGetter, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_address" {
			continue
		}
		_, err := client.Get(context.Background(), &vpc.GetAddressRequest{
			AddressId: rs.Primary.ID,
		})
		if status.Code(err) == codes.NotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("check yandex_vpc_address %q is deleted: %w", rs.Primary.ID, err)
		}
		return fmt.Errorf("address still exists: %s", rs.Primary.ID)
	}
	return nil
}
