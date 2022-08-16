package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func init() {
	resource.AddTestSweepers("yandex_vpc_gateway", &resource.Sweeper{
		Name: "yandex_vpc_gateway",
		F:    testSweepVPCGateways,
		Dependencies: []string{
			"yandex_vpc_route_table",
		},
	})
}

func testSweepVPCGateways(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListGatewaysRequest{FolderId: conf.FolderID}
	it := conf.sdk.VPC().Gateway().GatewayIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCGateway(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC gateway %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCGateway(conf *Config, id string) bool {
	return sweepWithRetry(sweepVPCGatewayOnce, conf, "VPC Gateway", id)
}

func sweepVPCGatewayOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCGatewayDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.VPC().Gateway().Delete(ctx, &vpc.DeleteGatewayRequest{
		GatewayId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccVPCGateway_basic(t *testing.T) {
	t.Parallel()

	var gateway vpc.Gateway
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	gatewayDesc := "Gateway description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCGateway_basic(gatewayName, gatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", gatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", gatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					testAccCheckVPCGatewayContainsLabel(&gateway, "tf-label", "tf-label-value"),
					testAccCheckVPCGatewayContainsLabel(&gateway, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_gateway.foo"),
				),
			},
			{
				ResourceName:      "yandex_vpc_gateway.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCGateway_update(t *testing.T) {
	t.Parallel()

	var gateway vpc.Gateway
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	gatewayDesc := "Gateway description for test"
	updatedGatewayName := gatewayName + "-update"
	updatedGatewayDesc := gatewayDesc + " with update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCGateway_basic(gatewayName, gatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", gatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", gatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_gateway.foo"),
				),
			},
			{
				Config: testAccVPCGateway_update(updatedGatewayName, updatedGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", updatedGatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", updatedGatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_gateway.foo"),
				),
			},
			{
				ResourceName:      "yandex_vpc_gateway.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCGatewayDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_gateway" {
			continue
		}

		_, err := config.sdk.VPC().Gateway().Get(context.Background(), &vpc.GetGatewayRequest{
			GatewayId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Gateway still exists")
		}
	}

	return nil
}

func testAccCheckVPCGatewayExists(n string, gateway *vpc.Gateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Gateway().Get(context.Background(), &vpc.GetGatewayRequest{
			GatewayId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Gateway not found")
		}

		*gateway = *found

		return nil
	}
}

func testAccCheckVPCGatewayContainsLabel(gateway *vpc.Gateway, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := gateway.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

//revive:disable:var-naming
func testAccVPCGateway_basic(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_gateway" "foo" {
  name        = "%s"
  description = "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  shared_egress_gateway {}
}
`, name, description)
}

func testAccVPCGateway_update(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_gateway" "foo" {
  name        = "%s"
  description = "%s"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }

  shared_egress_gateway {}
}
`, name, description)
}
