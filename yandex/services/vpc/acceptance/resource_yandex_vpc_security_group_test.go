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

func TestAccVPCSecurityGroup_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var securityGroup vpc.SecurityGroup

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	sg1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupBasic(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.to_port", "8099"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
				),
			},
			{
				ResourceName:      "yandex_vpc_security_group.sg1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroup_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var securityGroup vpc.SecurityGroup
	var securityGroup2 vpc.SecurityGroup

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	sg1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupBasic(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.to_port", "8099"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_security_group.sg1"),

					suite.testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &securityGroup2),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.port", "9000"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.predefined_target", "self_security_group"),
					// It's hard for test rule with security_group_id because of not stable hash of rule with ID.
					// predefined_target has the same logic. Assume that test covers this situation.
					statecheck.CheckCreatedAtAttr("yandex_vpc_security_group.sg2"),
				),
			},
			{
				Config: testAccVPCSecurityGroupBasic2(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "2"),
					// last rule can not be deleted because of optional+computed
					// setting ConfigModeAttr breaks json config format and migration
					// to terraform-plugin-framework
					// resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "0"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "ICMP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.to_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.from_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.to_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.from_port", "-1"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
				),
			},
			{
				ResourceName:            "yandex_vpc_security_group.sg1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"egress.0.id"},
			},
		},
	})
}

func (suite testSuite) testAccCheckVPCSecurityGroupExists(name string, securityGroup *vpc.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		client := vpcsdk.NewSecurityGroupClient(providerConfig.SDK)

		found, err := client.Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("security group not found")
		}

		*securityGroup = *found

		return nil
	}
}

func testAccVPCSecurityGroupBasic(networkName, sg1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sg1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "tcp"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    description    = "rule2 description"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}

resource "yandex_vpc_security_group" "sg2" {
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  egress {
    description       = "rule3 description"
    protocol          = "ANY"
    predefined_target = "self_security_group"
    port              = 9000
  }

  ingress {
    description       = "rule4 description"
    protocol          = "TCP"
    security_group_id = "${yandex_vpc_security_group.sg1.id}"
    port              = 9010
  }
}

`, networkName, sg1Name, acceptance.FolderID(), acceptance.FolderID())
}

func testAccVPCSecurityGroupBasic2(networkName, sg1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sg1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }

  ingress {
    description    = "rule1 description"
    protocol       = "icmp"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port = -1
  }

  ingress {
    description    = "rule2 description2"
    protocol       = "ANY"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
  }
}

`, networkName, sg1Name, acceptance.FolderID())
}

func (suite testSuite) testAccCheckVPCSecurityGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := vpcsdk.NewSecurityGroupClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_security_group" {
			continue
		}

		_, err := client.Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Security group still exists")
		}
	}

	return nil
}
