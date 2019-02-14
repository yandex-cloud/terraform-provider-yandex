package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccVPCSubnet_basic(t *testing.T) {
	t.Parallel()

	var subnet1 vpc.Subnet
	var subnet2 vpc.Subnet

	cnName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVPCSubnet_basic(cnName, subnet1Name, subnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-a", &subnet1),
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-b", &subnet2),
				),
			},
			resource.TestStep{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:      "yandex_vpc_subnet.subnet-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnet_basicV6(t *testing.T) {
	t.Skip("waiting ipv6 support in subnets")
	t.Parallel()

	var subnet1 vpc.Subnet
	var subnet2 vpc.Subnet

	cnName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVPCSubnet_basicV6(cnName, subnet1Name, subnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-a", &subnet1),
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-b", &subnet2),
				),
			},
			resource.TestStep{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:      "yandex_vpc_subnet.subnet-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_subnet" {
			continue
		}

		_, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
			SubnetId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Subnet still exists")
		}
	}

	return nil
}

func testAccCheckVPCSubnetExists(n string, subnet *vpc.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
			SubnetId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		*subnet = *found

		return nil
	}
}

//revive:disable:var-naming
func testAccVPCSubnet_basic(cnName, subnet1Name, subnet2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "custom-test" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.custom-test.id}"
  v4_cidr_blocks = ["10.0.0.0/16"]
}

resource "yandex_vpc_subnet" "subnet-b" {
  name           = "%s"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.custom-test.id}"
  v4_cidr_blocks = ["10.1.0.0/16"]
}
`, cnName, subnet1Name, subnet2Name)
}

func testAccVPCSubnet_basicV6(cnName, subnet1Name, subnet2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "custom-test" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.custom-test.id}"
  v4_cidr_blocks = ["10.0.0.0/16"]
  v6_cidr_blocks = ["fda9:8765:4321:1::/64"]
}

resource "yandex_vpc_subnet" "subnet-b" {
  name           = "%s"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.custom-test.id}"
  v4_cidr_blocks = ["10.1.0.0/16"]
  v6_cidr_blocks = ["fda9:8765:4321:2::/64"]
}
`, cnName, subnet1Name, subnet2Name)
}
