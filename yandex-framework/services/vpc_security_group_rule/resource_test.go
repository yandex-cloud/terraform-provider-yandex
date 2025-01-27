package vpc_security_group_rule_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	//testvpc "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/tests/vpc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const YandexVPCNetworkDefaultTimeout = 1 * time.Minute

func init() {
	resource.AddTestSweepers("yandex_vpc_security_group_rule", &resource.Sweeper{
		Name: "yandex_vpc_security_group_rule",
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

func TestAccVPCSecurityGroupRule_UpgradeFromSDKv2(t *testing.T) {
	networkName := acctest.RandomWithPrefix("vpc-sg-upgrade-provider")
	sgName := acctest.RandomWithPrefix("vpc-sg-upgrade-provider")

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.129.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testVPCSecurityGroupRuleBasicWithV4CidrTarget(networkName, sgName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(sgr1Name, "description", "hello there"),
					resource.TestCheckResourceAttr(sgr1Name, "direction", "ingress"),
					resource.TestCheckResourceAttr(sgr1Name, "port", "443"),
					resource.TestCheckResourceAttr(sgr1Name, "protocol", "TCP"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.#", "2"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.0", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(sgr1Name, "v4_cidr_blocks.1", "10.0.2.0/24"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testVPCSecurityGroupRuleBasicWithV4CidrTarget(networkName, sgName),
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

func TestAccVPCSecurityGroupRule_invalid(t *testing.T) {
	networkName := acctest.RandomWithPrefix("vpc-network")
	sgName := acctest.RandomWithPrefix("vpc-security-group")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testVPCSecurityGroupRuleInvalid(networkName, sgName),
				ExpectError: regexp.MustCompile("Use port attribute to specify single port value"),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_cidrBlocks(t *testing.T) {
	networkName := acctest.RandomWithPrefix("vpc-network")
	sgName := acctest.RandomWithPrefix("vpc-security-group")

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
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
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
			testVPCSecurityGroupRuleImportStep(sgr1Name, &sg1, &sgr1),
		},
	})
}

func TestAccVPCSecurityGroupRule_securityGroupId(t *testing.T) {
	networkName := acctest.RandomWithPrefix("vpc-network")
	sgName := acctest.RandomWithPrefix("vpc-security-group")
	sgName2 := acctest.RandomWithPrefix("vpc-security-group")

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sg2 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
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
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
			testVPCSecurityGroupRuleImportStep(sgr1Name, &sg1, &sgr1),
		},
	})
}

func TestAccVPCSecurityGroupRule_update(t *testing.T) {
	networkName := acctest.RandomWithPrefix("vpc-network")
	sgName := acctest.RandomWithPrefix("vpc-security-group")
	sgName2 := acctest.RandomWithPrefix("vpc-security-group")

	sgr1Name := "yandex_vpc_security_group_rule.sgr1"

	var sg1 vpc.SecurityGroup
	var sg2 vpc.SecurityGroup
	var sgr1 vpc.SecurityGroupRule

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckVPCSecurityGroupDestroy,
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
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
			testVPCSecurityGroupRuleImportStep(sgr1Name, &sg1, &sgr1),
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
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_id", sg2.GetId),
					test.AccCheckResourceAttrWithValueFactory(sgr1Name, "security_group_binding", sg1.GetId),
				),
			},
			testVPCSecurityGroupRuleImportStep(sgr1Name, &sg1, &sgr1),
		},
	})
}

func testVPCSecurityGroupRuleInvalid(networkName, sgName string) string {
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
  from_port              = 443
  to_port                = 443
  protocol               = "TCP"
}
`, networkName, sgName, test.GetExampleFolderID())
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
`, networkName, sgName, test.GetExampleFolderID())
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
`, networkName, sgName, test.GetExampleFolderID(), sgName2, test.GetExampleFolderID())
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
`, networkName, sgName, test.GetExampleFolderID(), sgName2, test.GetExampleFolderID())
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

func testVPCSecurityGroupRuleImportStep(resourceName string, securityGroup *vpc.SecurityGroup, securityGroupRule *vpc.SecurityGroupRule) resource.TestStep {
	return resource.TestStep{
		ResourceName: resourceName,
		ImportStateIdFunc: func(*terraform.State) (string, error) {
			return resourceid.Construct(securityGroup.Id, securityGroupRule.Id), nil
		},
		ImportState:       true,
		ImportStateVerify: true,
	}
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
