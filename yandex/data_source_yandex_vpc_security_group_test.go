package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func makeCheck(sg *vpc.SecurityGroup, folderID, name, desc string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceVPCSecurityGroupExists("data.yandex_vpc_security_group.sg1", sg),
		testAccCheckResourceIDField("data.yandex_vpc_security_group.sg1", "security_group_id"),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "name", name),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "description", desc),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "ingress.#", "1"),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
		resource.TestCheckResourceAttr("data.yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
		testAccCheckCreatedAtAttr("data.yandex_vpc_security_group.sg1"),
	)
}

func TestAccDataSourceVPCSecurityGroup_byID(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-sg")
	desc := "Description for test"
	folderID := getExampleFolderID()

	var sg vpc.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSecurityGroupConfig(name, desc, true),
				Check:  makeCheck(&sg, folderID, name, desc),
			},
			{
				Config: testAccDataSourceVPCSecurityGroupConfig(name, desc, false),
				Check:  makeCheck(&sg, folderID, name, desc),
			},
		},
	})
}

func TestAccDataSourceVPCSecurityGroup_byName(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-sg")
	desc := "Description for test"
	folderID := getExampleFolderID()

	var sg vpc.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSecurityGroupConfig(name, desc, false),
				Check:  makeCheck(&sg, folderID, name, desc),
			},
		},
	})
}

func testAccDataSourceVPCSecurityGroupExists(n string, sg *vpc.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().SecurityGroup().Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("security group not found")
		}

		*sg = *found

		return nil
	}
}

func testAccDataSourceVPCSecurityGroupConfig(name, desc string, useID bool) string {
	if useID {
		return testAccDataSourceVPCSecurityGroupResourceConfig(name, desc) + vpcSecurityGroupDataByIDConfig
	}

	return testAccDataSourceVPCSecurityGroupResourceConfig(name, desc) + vpcSecurityGroupDataByNameConfig
}

//revive:disable:var-naming
func testAccDataSourceVPCSecurityGroupResourceConfig(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net" {}

resource "yandex_vpc_security_group" "sg" {
  network_id  = "${yandex_vpc_network.net.id}"
  name        = "%s"
  description = "%s"
  ingress {
    description    = "rule1 description"
    protocol       = "TCP"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }
}
`, name, desc)
}

const vpcSecurityGroupDataByIDConfig = `
data "yandex_vpc_security_group" "sg1" {
  security_group_id = "${yandex_vpc_security_group.sg.id}"
}
`

const vpcSecurityGroupDataByNameConfig = `
data "yandex_vpc_security_group" "sg1" {
  name = "${yandex_vpc_security_group.sg.name}"
}
`
