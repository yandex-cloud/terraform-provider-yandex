package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func getYandexVPCSecurityGroupSweeperDeps() []string {
	return []string{
		"yandex_alb_load_balancer",
		"yandex_compute_instance",
		"yandex_compute_instance_group",
		"yandex_dataproc_cluster",
		"yandex_kubernetes_node_group",
		"yandex_kubernetes_cluster",
		"yandex_mdb_clickhouse_cluster",
		"yandex_mdb_mongodb_cluster",
		"yandex_mdb_mysql_cluster",
		"yandex_mdb_postgresql_cluster",
		"yandex_mdb_greenplum_cluster",
		"yandex_mdb_redis_cluster",
		"yandex_mdb_kafka_cluster",
		"yandex_mdb_sqlserver_cluster",
		"yandex_mdb_elasticsearch_cluster",
		"yandex_mdb_kafka_cluster",
	}
}

func init() {
	resource.AddTestSweepers("yandex_vpc_security_group", &resource.Sweeper{
		Name:         "yandex_vpc_security_group",
		F:            testSweepVPCSecurityGroups,
		Dependencies: getYandexVPCSecurityGroupSweeperDeps(),
	})
}

func testSweepVPCSecurityGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListSecurityGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.VPC().SecurityGroup().SecurityGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCSecurityGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC security group %q", it.Value().GetId()))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCSecurityGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepVPCSecurityGroupOnce, conf, "VPC Security Group", id)
}

func sweepVPCSecurityGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCNetworkDefaultTimeout)
	defer cancel()

	sg, err := conf.sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: id,
	})
	if err != nil {
		return err
	}

	if sg.DefaultForNetwork {
		return nil
	}

	op, err := conf.sdk.VPC().SecurityGroup().Delete(ctx, &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

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
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.to_port", "8099"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
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
	t.Parallel()

	var securityGroup vpc.SecurityGroup
	var securityGroup2 vpc.SecurityGroup

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
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.to_port", "8099"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),

					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &securityGroup2),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.port", "9000"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.predefined_target", "self_security_group"),
					// It's hard for test rule with security_group_id because of not stable hash of rule with ID.
					// predefined_target has the same logic. Assume that test covers this situation.
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sg2"),
				),
			},
			{
				Config: testAccVPCSecurityGroupBasic2(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg1", &securityGroup),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "2"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "ICMP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.to_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.from_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.to_port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.1.from_port", "-1"),
					testAccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
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

func testAccCheckVPCSecurityGroupExists(name string, securityGroup *vpc.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		sdk := testAccProvider.Meta().(*Config).sdk
		found, err := sdk.VPC().SecurityGroup().Get(context.Background(), &vpc.GetSecurityGroupRequest{
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

`, networkName, sg1Name, getExampleFolderID(), getExampleFolderID())
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

`, networkName, sg1Name, getExampleFolderID())
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
