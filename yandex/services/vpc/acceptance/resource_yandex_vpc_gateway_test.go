package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccVPCGateway_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var gateway vpc.Gateway
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	gatewayDesc := "Gateway description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCGateway_basic(gatewayName, gatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", gatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", gatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					testAccCheckVPCGatewayContainsLabel(&gateway, "tf-label", "tf-label-value"),
					testAccCheckVPCGatewayContainsLabel(&gateway, "empty-label", ""),
					statecheck.CheckCreatedAtAttr("yandex_vpc_gateway.foo"),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var gateway vpc.Gateway
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	gatewayDesc := "Gateway description for test"
	updatedGatewayName := gatewayName + "-update"
	updatedGatewayDesc := gatewayDesc + " with update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCGateway_basic(gatewayName, gatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", gatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", gatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_gateway.foo"),
				),
			},
			{
				Config: testAccVPCGateway_update(updatedGatewayName, updatedGatewayDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.foo", &gateway),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "name", updatedGatewayName),
					resource.TestCheckResourceAttr("yandex_vpc_gateway.foo", "description", updatedGatewayDesc),
					resource.TestCheckResourceAttrSet("yandex_vpc_gateway.foo", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_gateway.foo"),
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

func (suite testSuite) testAccCheckVPCGatewayDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_gateway" {
			continue
		}

		client := vpcsdk.NewGatewayClient(config.SDK)

		_, err := client.Get(context.Background(), &vpc.GetGatewayRequest{
			GatewayId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Gateway still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckVPCGatewayExists(n string, gateway *vpc.Gateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := vpcsdk.NewGatewayClient(config.SDK)

		found, err := client.Get(context.Background(), &vpc.GetGatewayRequest{
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
