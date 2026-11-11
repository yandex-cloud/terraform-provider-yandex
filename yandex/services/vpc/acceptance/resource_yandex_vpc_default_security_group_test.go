package acceptance

import (
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccVPCDefaultSecurityGroup_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := statecheck.RandomResourceName()

	var dsg vpc.SecurityGroup

	const dsgName = "yandex_vpc_default_security_group.dsg"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupBasic(networkName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists(dsgName, &dsg),
					statecheck.CheckResourceAttrWithValueFactory(dsgName, "name", dsg.GetName),
					resource.TestCheckResourceAttr(dsgName, "description", "hello there"),
					resource.TestCheckResourceAttr(dsgName, "ingress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "egress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.port", "-1"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.to_port", "8099"),
					resource.TestCheckResourceAttr(dsgName, "labels.empty-label", ""),
					resource.TestCheckResourceAttr(dsgName, "labels.tf-label", "my-tf-label"),
					statecheck.CheckCreatedAtAttr(dsgName),
				),
			},
			{
				ResourceName:      dsgName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := statecheck.RandomResourceName()

	var dsg vpc.SecurityGroup

	const dsgName = "yandex_vpc_default_security_group.dsg"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupBasic(networkName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists(dsgName, &dsg),
					statecheck.CheckResourceAttrWithValueFactory(dsgName, "name", dsg.GetName),
					resource.TestCheckResourceAttr(dsgName, "description", "hello there"),
					resource.TestCheckResourceAttr(dsgName, "ingress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "egress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.port", "-1"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.to_port", "8099"),
					resource.TestCheckResourceAttr(dsgName, "labels.%", "2"),
					resource.TestCheckResourceAttr(dsgName, "labels.empty-label", ""),
					resource.TestCheckResourceAttr(dsgName, "labels.tf-label", "my-tf-label"),
					statecheck.CheckCreatedAtAttr(dsgName),
				),
			},
			{
				Config: testAccVPCDefaultSecurityGroupBasic2(networkName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCSecurityGroupExists(dsgName, &dsg),
					statecheck.CheckResourceAttrWithValueFactory(dsgName, "name", dsg.GetName),
					resource.TestCheckResourceAttr(dsgName, "description", "updated description"),
					resource.TestCheckResourceAttr(dsgName, "ingress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "egress.#", "1"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(dsgName, "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.port", "-1"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr(dsgName, "egress.0.to_port", "8099"),
					resource.TestCheckResourceAttr(dsgName, "labels.%", "1"),
					resource.TestCheckResourceAttr(dsgName, "labels.new-label", "my-new-label"),
					statecheck.CheckCreatedAtAttr(dsgName),
				),
			},
			{
				ResourceName:      dsgName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_repeating(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := statecheck.RandomResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config:             testAccVPCDefaultSecurityGroupRepeating(networkName),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVPCDefaultSecurityGroupBasic(networkName string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network" {
	name = "%s"
}

resource "yandex_vpc_default_security_group" "dsg" {
	description = "hello there"
	network_id  = "${yandex_vpc_network.network.id}"
	folder_id   = "%s"

	labels = {
		tf-label = "my-tf-label"
		empty-label = ""
	}

	ingress {
		description    = "rule1 description"
		protocol       = "TCP"
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

`, networkName, acceptance.FolderID())
}

func testAccVPCDefaultSecurityGroupBasic2(networkName string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network" {
	name = "%s"
}

resource "yandex_vpc_default_security_group" "dsg" {
	description = "updated description"
	network_id  = "${yandex_vpc_network.network.id}"
	folder_id   = "%s"

	labels = {
		new-label = "my-new-label"
	}

	ingress {
		description    = "rule1 description"
		protocol       = "TCP"
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

`, networkName, acceptance.FolderID())
}

func testAccVPCDefaultSecurityGroupRepeating(networkName string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network" {
	name = "%s"
}

resource "yandex_vpc_default_security_group" "dsg" {
	description = "hello there"
	network_id  = "${yandex_vpc_network.network.id}"
	folder_id   = "%s"

	labels = {
		tf-label = "my-tf-label"
		empty-label = ""
	}

	ingress {
		description    = "rule1 description"
		protocol       = "TCP"
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

resource "yandex_vpc_default_security_group" "same-dsg" {
	description = "updated description"
	network_id  = "${yandex_vpc_network.network.id}"
	folder_id   = "%s"

	labels = {
		new-label = "my-new-label"
	}

	ingress {
		description    = "rule1 description"
		protocol       = "TCP"
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

`, networkName, acceptance.FolderID(), acceptance.FolderID())
}
