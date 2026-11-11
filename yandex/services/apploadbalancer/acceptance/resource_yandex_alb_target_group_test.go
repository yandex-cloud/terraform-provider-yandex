package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	albsdk "github.com/yandex-cloud/go-sdk/services/apploadbalancer/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	computefixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/testfixtures"
)

const albTGResource = "yandex_alb_target_group.test-tg"

func albTargetGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albTGResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBTargetGroup_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					resource.TestCheckResourceAttr(albTGResource, "name", tgName),
					resource.TestCheckResourceAttrSet(albTGResource, "folder_id"),
					resource.TestCheckResourceAttr(albTGResource, "folder_id", folderID),
					testAccCheckALBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckALBTargetGroupContainsLabel(&tg, "empty-label", ""),
					statecheck.CheckCreatedAtAttr(albTGResource),
					suite.testAccCheckALBTargetGroupValues(&tg, []string{}),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func TestAccALBTargetGroup_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccALBBaseTemplate(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					suite.testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					statecheck.ExistsFirstElementWithAttr(
						albTGResource, "target", "subnet_id", &targetPath,
					),
					statecheck.CheckResourceSubAttrFn(
						albTGResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					statecheck.CheckResourceSubAttrFn(
						albTGResource, &targetPath, "ip_address", func(value string) error {
							address := tg.GetTargets()[0].GetIpAddress()
							if value != address {
								return fmt.Errorf("TargetGroup's target's address doesnt't match. %s != %s", value, address)
							}
							return nil
						},
					),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func TestAccALBTargetGroup_externalAddress(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group-ext")
	tgDesc := "Description for test"
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBTargetGroupExternalAddress(tgName, tgDesc, testAccALBBaseTemplate(instancePrefix)),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					resource.TestCheckResourceAttr(albTGResource, "name", tgName),
					resource.TestCheckResourceAttr(albTGResource, "target.#", "1"),
					resource.TestCheckResourceAttr(albTGResource, "target.0.external_address", "true"),
					testAccCheckALBTargetGroupExternalAddressTarget(&tg, true),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func TestAccALBTargetGroup_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var tg apploadbalancer.TargetGroup
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccALBBaseTemplate(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					suite.testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", computefixtures.TwoInstancesWithSubnet(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					suite.testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", computefixtures.TwoInstancesWithSubnet(instancePrefix), 2, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBTargetGroupExists(albTGResource, &tg),
					suite.testAccCheckALBTargetGroupValues(&tg, []string{
						fmt.Sprintf("%s-1", instancePrefix), fmt.Sprintf("%s-2", instancePrefix),
					}),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func (suite testSuite) testAccCheckALBTargetGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := albsdk.NewTargetGroupClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_target_group" {
			continue
		}

		_, err := client.Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckALBTargetGroupExists(tgName string, tg *apploadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tgName]
		if !ok {
			return fmt.Errorf("Not found: %s", tgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := albsdk.NewTargetGroupClient(config.SDK)

		found, err := client.Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("TargetGroup not found")
		}

		*tg = *found

		return nil
	}
}

func testAccCheckALBTargetGroupExternalAddressTarget(tg *apploadbalancer.TargetGroup, expectExternal bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		targets := tg.GetTargets()
		if len(targets) != 1 {
			return fmt.Errorf("expected 1 target in target group, got %d", len(targets))
		}
		t := targets[0]
		if t.GetSubnetId() != "" {
			return fmt.Errorf("expected empty subnet_id for external_address target, got %q", t.GetSubnetId())
		}
		if t.GetExternalAddress() != expectExternal {
			return fmt.Errorf("expected external_address=%v from API, got %v", expectExternal, t.GetExternalAddress())
		}
		if t.GetIpAddress() == "" {
			return fmt.Errorf("expected non-empty target ip_address from API")
		}
		return nil
	}
}

func testAccALBTargetGroupExternalAddress(tgName, tgDesc, baseTemplate string) string {
	return fmt.Sprintf(`
%s

resource "yandex_alb_target_group" "test-tg" {
  name = "%s"
  description = "%s"

  target {
    ip_address       = yandex_compute_instance.test-instance-1.network_interface.0.ip_address
    external_address = true
  }
}
`, baseTemplate, tgName, tgDesc)
}

func testAccCheckALBTargetGroupContainsLabel(tg *apploadbalancer.TargetGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := tg.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccALBTargetGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_alb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}
