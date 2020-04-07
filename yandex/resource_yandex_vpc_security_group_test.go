package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccVPCSecurityGroup_basic(t *testing.T) {
	t.Parallel()

	var securityGroup vpc.SecurityGroup

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	sg1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupBasic(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists(
						"yandex_vpc_security_group.sgr1", &securityGroup),

					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.direction", "EGRESS"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sgr1"),
				),
			},
			{
				ResourceName:      "yandex_vpc_security_group.sgr1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroup_update(t *testing.T) {
	t.Parallel()

	var securityGroup vpc.SecurityGroup

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	sg1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupBasic(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists(
						"yandex_vpc_security_group.sgr1", &securityGroup),

					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.#", "2"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.direction", "EGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.port", "0"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.from_port", "8090"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sgr1"),
				),
			},
			{
				Config: testAccVPCSecurityGroupBasic2(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists(
						"yandex_vpc_security_group.sgr1", &securityGroup),

					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.#", "2"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.port", "0"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.from_port", "8091"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sgr1"),
				),
			},
			{
				Config: testAccVPCSecurityGroupBasic3(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists(
						"yandex_vpc_security_group.sgr1", &securityGroup),

					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.#", "3"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.direction", "INGRESS"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.port", "0"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sgr1", "rule.1.from_port", "8091"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sgr1"),
				),
			},
			{
				ResourceName:      "yandex_vpc_security_group.sgr1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCSecurityGroupExists(name string, securityGroup *vpc.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		sdk := testAccProvider.Meta().(*Config).sdk
		found, err := sdk.VPC().SecurityGroup().Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Security group not found")
		}

		*securityGroup = *found

		return nil
	}
}

func testAccVPCSecurityGroupBasic(networkName, sgr1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sgr1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  rule {
    direction = "INGRESS"
    description = "rule1 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    port = 8080
  }

  rule {
    direction = "EGRESS"
    description = "rule2 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    from_port = 8090
    to_port = 8099
  }
}

`, networkName, getExampleFolderID(), sgr1Name)
}

func testAccVPCSecurityGroupBasic2(networkName, sgr1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sgr1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  rule {
    direction = "INGRESS"
    description = "rule1 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    port = 8080
  }

  rule {
    direction = "INGRESS"
    description = "rule2 description2"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    from_port = 8091
    to_port = 8099
  }
}

`, networkName, getExampleFolderID(), sgr1Name)
}

func testAccVPCSecurityGroupBasic3(networkName, sgr1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sgr1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  rule {
    direction = "INGRESS"
    description = "rule1 description"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    port = 8080
  }

  rule {
    direction = "INGRESS"
    description = "rule2 description2"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    from_port = 8091
    to_port = 8099
  }

  rule {
    direction = "INGRESS"
    description = "rule3 description2"
	v4_cidr_blocks = ["10.0.0.1/24", "10.0.0.2/24"]
    port = 9999
  }
}

`, networkName, getExampleFolderID(), sgr1Name)
}

func testAccCheckVPCSecurityGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_security_group" {
			continue
		}

		_, err := config.sdk.VPC().SecurityGroup().Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Security group still exists")
		}
	}

	return nil
}
