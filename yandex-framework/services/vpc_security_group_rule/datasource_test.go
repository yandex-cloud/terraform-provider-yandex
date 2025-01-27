package vpc_security_group_rule_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceVPCSecurityGroupRule(t *testing.T) {
	sgr1Name := "data.yandex_vpc_security_group_rule.rule1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSecurityGroupRuleResourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sgr1Name, "description", "rule1 description"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "ingress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "8080"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "TCP"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.0", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.1", "10.0.2.0/24"),
				),
			},
		},
	})
}

func testAccDataSourceVPCSecurityGroupRuleResourceConfig() string {
	return `
resource "yandex_vpc_network" "net" {}

resource "yandex_vpc_security_group" "sg" {
  network_id  = "${yandex_vpc_network.net.id}"
  name        = "some-name"
  description = "some description"
}

resource "yandex_vpc_security_group_rule" "rule" {
  security_group_binding = "${yandex_vpc_security_group.sg.id}"
  description    = "rule1 description"
  direction      = "ingress"
  protocol       = "TCP"
  v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
  port           = 8080
}

data "yandex_vpc_security_group_rule" "rule1" {
  security_group_binding = "${yandex_vpc_security_group.sg.id}"
  rule_id                = "${yandex_vpc_security_group_rule.rule.id}"
}
`
}
