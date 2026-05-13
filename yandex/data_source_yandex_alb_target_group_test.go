package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albTgDataSourceResource = "data.yandex_alb_target_group.test-tg-ds"

// albTgPrivateExternalTestIPAddress is a fixed private IPv4 used in acc tests for
// external_address targets without provisioning compute.
const albTgPrivateExternalTestIPAddress = "192.168.10.37"

func TestAccDataSourceALBTargetGroup_byID(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	folderID := getExampleFolderID()

	var tg apploadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBTargetGroupConfigByID(tgName, tgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBTargetGroupExists(albTgDataSourceResource, &tg),
					testAccCheckResourceIDField(albTgDataSourceResource, "target_group_id"),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "name", tgName),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "description", tgDesc),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albTgDataSourceResource),
					testAccCheckALBTargetGroupValues(&tg, []string{}),
				),
			},
		},
	})
}

func TestAccDataSourceALBTargetGroup_byName(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	folderID := getExampleFolderID()

	var tg apploadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBTargetGroupConfigByName(tgName, tgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBTargetGroupExists(albTgDataSourceResource, &tg),
					testAccCheckResourceIDField(albTgDataSourceResource, "target_group_id"),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "name", tgName),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "description", tgDesc),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albTgDataSourceResource),
					testAccCheckALBTargetGroupValues(&tg, []string{}),
				),
			},
		},
	})
}

func TestAccDataSourceALBTargetGroup_full(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")
	var tg apploadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralTGTemplate(tgName, tgDesc, testAccALBBaseTemplate(instancePrefix), 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBTargetGroupExists(albTgDataSourceResource, &tg),
					testAccCheckALBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.0.private_ipv4_address", "false"),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.0.external_address", "false"),
					testExistsFirstElementWithAttr(
						albTgDataSourceResource, "target", "subnet_id", &targetPath,
					),
					testCheckResourceSubAttrFn(
						albTgDataSourceResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albTgDataSourceResource, &targetPath, "ip_address", func(value string) error {
							address := tg.GetTargets()[0].GetIpAddress()
							if value != address {
								return fmt.Errorf("TargetGroup's target's address doesnt't match. %s != %s", value, address)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBTargetGroup_privateAndExternalAddress(t *testing.T) {
	t.Parallel()

	var tg apploadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBTargetGroupPrivateAndExternal(tgName, tgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBTargetGroupExists(albTgDataSourceResource, &tg),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "name", tgName),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.#", "1"),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.0.ip_address", albTgPrivateExternalTestIPAddress),
					resource.TestCheckResourceAttr(albTgDataSourceResource, "target.0.external_address", "true"),
					testAccCheckALBTargetGroupPrivateIPv4AndExternalTargets(&tg, true),
				),
			},
		},
	})
}

func testAccCheckALBTargetGroupPrivateIPv4AndExternalTargets(tg *apploadbalancer.TargetGroup, wantSubnetEmpty bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ts := tg.GetTargets()
		if len(ts) != 1 {
			return fmt.Errorf("expected 1 target, got %d", len(ts))
		}
		t := ts[0]
		if wantSubnetEmpty && t.GetSubnetId() != "" {
			return fmt.Errorf("expected empty subnet_id, got %q", t.GetSubnetId())
		}
		if !t.GetExternalAddress() {
			return fmt.Errorf("expected external_address true from API, got false")
		}
		if t.GetIpAddress() == "" {
			return fmt.Errorf("expected non-empty ip_address from API")
		}
		return nil
	}
}

func testAccDataSourceALBTargetGroupExists(n string, tg *apploadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(context.Background(), &apploadbalancer.GetTargetGroupRequest{
			TargetGroupId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("TargetGroup not found")
		}

		*tg = *found

		return nil
	}
}

func testAccDataSourceALBTargetGroupConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_target_group" "test-tg-ds" {
  target_group_id = "${yandex_alb_target_group.test-tg.id}"
}

resource "yandex_alb_target_group" "test-tg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceALBTargetGroupConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_target_group" "test-tg-ds" {
  name = "${yandex_alb_target_group.test-tg.name}"
}

resource "yandex_alb_target_group" "test-tg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceALBTargetGroupPrivateAndExternal(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_target_group" "test-tg-ds" {
  name = yandex_alb_target_group.test-tg.name
}

resource "yandex_alb_target_group" "test-tg" {
  name        = "%s"
  description = "%s"

  target {
    ip_address           = "%s"
    external_address     = true
  }
}
`, name, desc, albTgPrivateExternalTestIPAddress)
}
