package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func init() {
	resource.AddTestSweepers("yandex_vpc_security_group_rule", &resource.Sweeper{
		Name:         "yandex_vpc_security_group_rule",
		F:            testSweepVPCSecurityGroups,
		Dependencies: append(getYandexVPCSecurityGroupSweeperDeps(), "yandex_vpc_security_group"),
	})
}

func TestAccVPCSecurityGroupRule_cidrBlocks(t *testing.T) {
	t.Parallel()

	networkName := getRandAccTestResourceName()
	sgName := getRandAccTestResourceName()

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testVPCSecurityGroupRuleBasicWithV4CidrTarget(networkName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &sg1),
					testAccCheckVPCSecurityGroupRuleExists(sgr1Name, &sg1, &sgr1),
					resource.TestCheckResourceAttr(sgr1Name, "description", "hello there"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "ingress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "443"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "TCP"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.0", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.1", "10.0.2.0/24"),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_securityGroupId(t *testing.T) {
	t.Parallel()

	networkName := getRandAccTestResourceName()
	sgName := getRandAccTestResourceName()
	sgName2 := getRandAccTestResourceName()

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sg2 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testVPCSecurityGroupRuleBasicWithSecurityGroupTarget(networkName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &sg1),
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &sg2),
					testAccCheckVPCSecurityGroupRuleExists(sgr1Name, &sg1, &sgr1),
					resource.TestCheckResourceAttr(sgr1Name, "description", "hello there"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "egress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "31337"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "UDP"),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_update(t *testing.T) {
	t.Parallel()

	networkName := getRandAccTestResourceName()
	sgName := getRandAccTestResourceName()
	sgName2 := getRandAccTestResourceName()

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sg2 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testVPCSecurityGroupRuleBasicWithSecurityGroupTarget(networkName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &sg1),
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &sg2),
					testAccCheckVPCSecurityGroupRuleExists(sgr1Name, &sg1, &sgr1),
					resource.TestCheckResourceAttr(sgr1Name, "description", "hello there"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "egress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "31337"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "UDP"),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
			{
				Config: testVPCSecurityGroupRuleBasicWithSecurityGroupTarget_updated(networkName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &sg1),
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &sg2),
					testAccCheckVPCSecurityGroupRuleExists(sgr1Name, &sg1, &sgr1),
					resource.TestCheckResourceAttr(sgr1Name, "description", "hello there"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "ingress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "1337"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "UDP"),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					testAccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
		},
	})
}

func testVPCSecurityGroupRuleBasicWithV4CidrTarget(networkName, sgName string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sg1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"
}

resource "yandex_vpc_security_group_rule" "sgr1" {
  description            = "hello there"
  direction              = "ingress"
  v4_cidr_blocks         = ["10.0.1.0/24", "10.0.2.0/24"]
  security_group_binding = yandex_vpc_security_group.sg1.id
  port                   = 443
  protocol               = "TCP"
}
`, networkName, sgName, getExampleFolderID())
}

func testVPCSecurityGroupRuleBasicWithSecurityGroupTarget_updated(networkName, sgName, sgName2 string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sg1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"
}

resource "yandex_vpc_security_group" "sg2" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"
}

resource "yandex_vpc_security_group_rule" "sgr1" {
  description            = "hello there"
  direction              = "ingress"
  security_group_id      = yandex_vpc_security_group.sg2.id
  security_group_binding = yandex_vpc_security_group.sg1.id
  port                   = 1337
  protocol               = "UDP"
}
`, networkName, sgName, getExampleFolderID(), sgName2, getExampleFolderID())
}

func testVPCSecurityGroupRuleBasicWithSecurityGroupTarget(networkName, sgName, sgName2 string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_security_group" "sg1" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"
}

resource "yandex_vpc_security_group" "sg2" {
  name        = "%s"
  description = "description for security group"
  network_id  = "${yandex_vpc_network.foo.id}"
  folder_id   = "%s"
}

resource "yandex_vpc_security_group_rule" "sgr1" {
  description            = "hello there"
  direction              = "egress"
  security_group_id      = yandex_vpc_security_group.sg2.id
  security_group_binding = yandex_vpc_security_group.sg1.id
  port                   = 31337
  protocol               = "UDP"
}
`, networkName, sgName, getExampleFolderID(), sgName2, getExampleFolderID())
}

func testAccCheckVPCSecurityGroupRuleExists(name string, securityGroup *vpc.SecurityGroup, securityGroupRule *vpc.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		ruleId := rs.Primary.ID

		for i := range securityGroup.Rules {
			rule := *securityGroup.Rules[i]

			if rule.Id == ruleId {
				*securityGroupRule = rule

				return nil
			}
		}

		return fmt.Errorf("security group rule not found")
	}
}
