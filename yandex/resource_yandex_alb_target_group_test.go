package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
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
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					resource.TestCheckResourceAttr(albTGResource, "name", tgName),
					resource.TestCheckResourceAttrSet(albTGResource, "folder_id"),
					resource.TestCheckResourceAttr(albTGResource, "folder_id", folderID),
					testAccCheckALBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckALBTargetGroupContainsLabel(&tg, "empty-label", ""),
					testAccCheckCreatedAtAttr(albTGResource),
					testAccCheckALBTargetGroupValues(&tg, []string{}),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func TestAccALBTargetGroup_full(t *testing.T) {
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccALBBaseTemplate(instancePrefix), 1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					testExistsFirstElementWithAttr(
						albTGResource, "target", "subnet_id", &targetPath,
					),
					testCheckResourceSubAttrFn(
						albTGResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
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

func TestAccALBTargetGroup_update(t *testing.T) {
	var tg apploadbalancer.TargetGroup
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group", "tf-descr", testAccALBBaseTemplate(instancePrefix), 1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", testAccLBBaseTemplate(instancePrefix), 1,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccALBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", testAccLBBaseTemplate(instancePrefix), 2,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBTargetGroupExists(albTGResource, &tg),
					testAccCheckALBTargetGroupValues(&tg, []string{
						fmt.Sprintf("%s-1", instancePrefix), fmt.Sprintf("%s-2", instancePrefix),
					}),
				),
			},
			albTargetGroupImportStep(),
		},
	})
}

func testAccCheckALBTargetGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_target_group" {
			continue
		}

		_, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func testAccCheckALBTargetGroupExists(tgName string, tg *apploadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tgName]
		if !ok {
			return fmt.Errorf("Not found: %s", tgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
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
