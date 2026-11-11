package acceptance

import (
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/testfixtures"
)

func testAccDataSourceVPCAddressConfig(name string, useID bool) string {
	if useID {
		return testAccDataSourceVPCAddressResourceConfig(name) + vpcAddressDataByIDConfig
	}

	return testAccDataSourceVPCAddressResourceConfig(name) + vpcAddressDataByNameConfig
}

func testAccDataSourceVPCAddressResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_address" "addr" {
  name = "%s"

  external_ipv4_address {
    zone_id = "ru-central1-d"
  }
}
`, name)
}

func testAccDataSourceVPCAddressInternalConfig(name string, useID bool) string {
	if useID {
		return testAccDataSourceVPCAddressInternalResourceConfig(name) + vpcAddressDataByIDConfig
	}

	return testAccDataSourceVPCAddressInternalResourceConfig(name) + vpcAddressDataByNameConfig
}

func testAccDataSourceVPCAddressInternalResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr" {
  name = "%[1]s"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
}
`, name)
}

const vpcAddressDataByIDConfig = `
data "yandex_vpc_address" "addr1" {
  address_id = "${yandex_vpc_address.addr.id}"
}
`

const vpcAddressDataByNameConfig = `
data "yandex_vpc_address" "addr1" {
  name = "${yandex_vpc_address.addr.name}"
}
`

func TestAccDataSourceVPCAddress_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	addressName := acctest.RandomWithPrefix("tf-address")
	folderID := acceptance.FolderID()

	var address vpc.Address

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: testfixtures.CheckAddressDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCAddressConfig(addressName, true),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCAddressExists("data.yandex_vpc_address.addr1", &address),
					statecheck.CheckResourceIDField("data.yandex_vpc_address.addr1", "address_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "name", addressName),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "deletion_protection", "false"),
					statecheck.CheckCreatedAtAttr("data.yandex_vpc_address.addr1"),
				),
			},
			{
				Config: testAccDataSourceVPCAddressConfig(addressName, false),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCAddressExists("data.yandex_vpc_address.addr1", &address),
					statecheck.CheckResourceIDField("data.yandex_vpc_address.addr1", "address_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "name", addressName),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "folder_id", folderID),
					statecheck.CheckCreatedAtAttr("data.yandex_vpc_address.addr1"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCAddress_internal(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	addressName := acctest.RandomWithPrefix("tf-address")
	folderID := acceptance.FolderID()

	var address vpc.Address

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: testfixtures.CheckAddressDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCAddressInternalConfig(addressName, true),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCAddressExists("data.yandex_vpc_address.addr1", &address),
					statecheck.CheckResourceIDField("data.yandex_vpc_address.addr1", "address_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "name", addressName),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "internal_ipv4_address.#", "1"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_address.addr1", "internal_ipv4_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_address.addr1", "internal_ipv4_address.0.address"),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "deletion_protection", "false"),
					statecheck.CheckCreatedAtAttr("data.yandex_vpc_address.addr1"),
				),
			},
			{
				Config: testAccDataSourceVPCAddressInternalConfig(addressName, false),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCAddressExists("data.yandex_vpc_address.addr1", &address),
					statecheck.CheckResourceIDField("data.yandex_vpc_address.addr1", "address_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "name", addressName),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_address.addr1", "internal_ipv4_address.#", "1"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_address.addr1", "internal_ipv4_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_address.addr1", "internal_ipv4_address.0.address"),
					statecheck.CheckCreatedAtAttr("data.yandex_vpc_address.addr1"),
				),
			},
		},
	})
}
