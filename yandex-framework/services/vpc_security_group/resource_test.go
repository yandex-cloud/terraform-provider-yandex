package vpc_security_group_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	//testvpc "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/tests/vpc"
)

const YandexVPCNetworkDefaultTimeout = 1 * time.Minute

func init() {
	resource.AddTestSweepers("yandex_vpc_security_group", &resource.Sweeper{
		Name: "yandex_vpc_security_group",
		F:    testSweepVPCSecurityGroups,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepVPCSecurityGroups(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListSecurityGroupsRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	it := conf.SDK.VPC().SecurityGroup().SecurityGroupIterator(context.Background(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCSecurityGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC security group %q", it.Value().GetId()))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCSecurityGroup(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepVPCSecurityGroupOnce, conf, "VPC Security Group", id)
}

func sweepVPCSecurityGroupOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), YandexVPCNetworkDefaultTimeout)
	defer cancel()

	sg, err := conf.SDK.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: id,
	})
	if err != nil {
		return err
	}

	if sg.DefaultForNetwork {
		return nil
	}

	op, err := conf.SDK.VPC().SecurityGroup().Delete(ctx, &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

func TestAccVPCSecurityGroup_UpgradeFromSDKv2(t *testing.T) {
	t.Skip()

	networkName := acctest.RandomWithPrefix("vpc-sg-upgrade-provider")
	sg1Name := acctest.RandomWithPrefix("vpc-sg-upgrade-provider")

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.129.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccVPCSecurityGroupBasic(networkName, sg1Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "ingress.0.port", "8080"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.port", "-1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.from_port", "8090"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg1", "egress.0.to_port", "8099"),
					test.AccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccVPCSecurityGroupBasic(networkName, sg1Name),
				// ConfigPlanChecks is a terraform-plugin-testing feature.
				// If acceptance testing is still using terraform-plugin-sdk/v2,
				// use `PlanOnly: true` instead. When migrating to
				// terraform-plugin-testing, switch to `ConfigPlanChecks` or you
				// will likely experience test failures.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccVPCSecurityGroup_basic(t *testing.T) {
	t.Skip()

	var securityGroup vpc.SecurityGroup

	networkName := acctest.RandomWithPrefix("vpc-sg-basic")
	sg1Name := acctest.RandomWithPrefix("vpc-sg-basic")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
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
					test.AccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
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
	t.Skip()

	var securityGroup vpc.SecurityGroup
	var securityGroup2 vpc.SecurityGroup

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	sg1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
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
					test.AccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),

					testAccCheckVPCSecurityGroupExists("yandex_vpc_security_group.sg2", &securityGroup2),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "ingress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.protocol", "ANY"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.port", "9000"),
					resource.TestCheckResourceAttr("yandex_vpc_security_group.sg2", "egress.0.predefined_target", "self_security_group"),
					// It's hard for test rule with security_group_id because of not stable hash of rule with ID.
					// predefined_target has the same logic. Assume that test covers this situation.
					test.AccCheckCreatedAtAttr("yandex_vpc_security_group.sg2"),
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
					test.AccCheckCreatedAtAttr("yandex_vpc_security_group.sg1"),
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

		sdk := test.AccProvider.(*yandex_framework.Provider).GetConfig().SDK
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

`, networkName, sg1Name, test.GetExampleFolderID(), test.GetExampleFolderID())
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

`, networkName, sg1Name, test.GetExampleFolderID())
}

func testAccCheckVPCSecurityGroupDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_security_group" {
			continue
		}

		_, err := config.SDK.VPC().SecurityGroup().Get(context.Background(), &vpc.GetSecurityGroupRequest{
			SecurityGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Security group still exists")
		}
	}

	return nil
}
