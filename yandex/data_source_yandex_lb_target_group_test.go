package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
)

const tgDataSourceResource = "data.yandex_lb_target_group.test-tg-ds"

func TestAccDataSourceLBTargetGroup_byID(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	folderID := getExampleFolderID()

	var tg loadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBTargetGroupConfigByID(tgName, tgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBTargetGroupExists(tgDataSourceResource, &tg),
					testAccCheckResourceIDField(tgDataSourceResource, "target_group_id"),
					resource.TestCheckResourceAttr(tgDataSourceResource, "name", tgName),
					resource.TestCheckResourceAttr(tgDataSourceResource, "description", tgDesc),
					resource.TestCheckResourceAttr(tgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(tgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(tgDataSourceResource),
					testAccCheckLBTargetGroupValues(&tg, []string{}),
				),
			},
		},
	})
}

func TestAccDataSourceLBTargetGroup_byName(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	folderID := getExampleFolderID()

	var tg loadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceLBTargetGroupConfigByName(tgName, tgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBTargetGroupExists(tgDataSourceResource, &tg),
					testAccCheckResourceIDField(tgDataSourceResource, "target_group_id"),
					resource.TestCheckResourceAttr(tgDataSourceResource, "name", tgName),
					resource.TestCheckResourceAttr(tgDataSourceResource, "description", tgDesc),
					resource.TestCheckResourceAttr(tgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(tgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(tgDataSourceResource),
					testAccCheckLBTargetGroupValues(&tg, []string{}),
				),
			},
		},
	})
}

func TestAccDataSourceLBTargetGroup_full(t *testing.T) {
	t.Parallel()

	tgName := acctest.RandomWithPrefix("tf-tg")
	tgDesc := "Description for test"
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")
	var tg loadbalancer.TargetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralTGTemplate(tgName, tgDesc, testAccLBBaseTemplate(instancePrefix), 1, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceLBTargetGroupExists(tgDataSourceResource, &tg),
					testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					testExistsFirstElementWithAttr(
						tgDataSourceResource, "target", "subnet_id", &targetPath,
					),
					testCheckResourceSubAttrFn(
						tgDataSourceResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						tgDataSourceResource, &targetPath, "address", func(value string) error {
							address := tg.GetTargets()[0].Address
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

func testAccDataSourceLBTargetGroupExists(n string, tg *loadbalancer.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LoadBalancer().TargetGroup().Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
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

func testAccDataSourceLBTargetGroupConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_lb_target_group" "test-tg-ds" {
  target_group_id = "${yandex_lb_target_group.test-tg.id}"
}

resource "yandex_lb_target_group" "test-tg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceLBTargetGroupConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_lb_target_group" "test-tg-ds" {
  name = "${yandex_lb_target_group.test-tg.name}"
}

resource "yandex_lb_target_group" "test-tg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}
