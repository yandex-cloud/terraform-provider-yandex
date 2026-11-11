package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	loadbalancersdk "github.com/yandex-cloud/go-sdk/services/loadbalancer/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	computefixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/testfixtures"
)

const tgResource = "yandex_lb_target_group.test-tg"

func targetGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      tgResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccLBTargetGroup_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg loadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBTargetGroupBasic(tgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					resource.TestCheckResourceAttr(tgResource, "name", tgName),
					resource.TestCheckResourceAttrSet(tgResource, "folder_id"),
					resource.TestCheckResourceAttr(tgResource, "folder_id", folderID),
					testAccCheckLBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckLBTargetGroupContainsLabel(&tg, "empty-label", ""),
					statecheck.CheckCreatedAtAttr(tgResource),
					suite.testAccCheckLBTargetGroupValues(&tg, []string{}),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func TestAccLBTargetGroup_dynamicBlock(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg loadbalancer.TargetGroup
	tgName := acctest.RandomWithPrefix("tf-target-group")
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBTargetGroupDynamicBlock(tgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					resource.TestCheckResourceAttr(tgResource, "name", tgName),
					resource.TestCheckResourceAttrSet(tgResource, "folder_id"),
					resource.TestCheckResourceAttr(tgResource, "folder_id", folderID),
					testAccCheckLBTargetGroupContainsLabel(&tg, "tf-label", "tf-label-value"),
					testAccCheckLBTargetGroupContainsLabel(&tg, "empty-label", ""),
					statecheck.CheckCreatedAtAttr(tgResource),
					//testAccCheckLBTargetGroupValues(&tg, []string{}),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func TestAccLBTargetGroup_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var tg loadbalancer.TargetGroup
	targetPath := ""
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group", "tf-descr", computefixtures.TwoInstancesWithSubnet(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					suite.testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
					statecheck.ExistsFirstElementWithAttr(
						tgResource, "target", "subnet_id", &targetPath,
					),
					statecheck.CheckResourceSubAttrFn(
						tgResource, &targetPath, "subnet_id", func(value string) error {
							subnetID := tg.GetTargets()[0].SubnetId
							if value != subnetID {
								return fmt.Errorf("TargetGroup's target's sudnet_id doesnt't match. %s != %s", value, subnetID)
							}
							return nil
						},
					),
					statecheck.CheckResourceSubAttrFn(
						tgResource, &targetPath, "address", func(value string) error {
							address := tg.GetTargets()[0].Address
							if value != address {
								return fmt.Errorf("TargetGroup's target's address doesnt't match. %s != %s", value, address)
							}
							return nil
						},
					),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func TestAccLBTargetGroup_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var tg loadbalancer.TargetGroup
	instancePrefix := acctest.RandomWithPrefix("tf-instance")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckLBNetworkLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group", "tf-descr", computefixtures.TwoInstancesWithSubnet(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					suite.testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", computefixtures.TwoInstancesWithSubnet(instancePrefix), 1, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					suite.testAccCheckLBTargetGroupValues(&tg, []string{fmt.Sprintf("%s-1", instancePrefix)}),
				),
			},
			{
				Config: testAccLBGeneralTGTemplate(
					"tf-target-group-updated", "tf-descr-updated", computefixtures.TwoInstancesWithSubnet(instancePrefix), 2, false,
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLBTargetGroupExists(tgResource, &tg),
					suite.testAccCheckLBTargetGroupValues(&tg, []string{
						fmt.Sprintf("%s-1", instancePrefix), fmt.Sprintf("%s-2", instancePrefix),
					}),
				),
			},
			targetGroupImportStep(),
		},
	})
}

func (suite testSuite) testAccCheckLBTargetGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := loadbalancersdk.NewTargetGroupClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_lb_target_group" {
			continue
		}

		_, err := client.Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
			TargetGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("TargetGroup still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckLBTargetGroupExists(tgName string, tg *loadbalancer.TargetGroup) resource.TestCheckFunc {
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

		client := loadbalancersdk.NewTargetGroupClient(config.SDK)

		found, err := client.Get(context.Background(), &loadbalancer.GetTargetGroupRequest{
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

func testAccCheckLBTargetGroupContainsLabel(tg *loadbalancer.TargetGroup, key string, value string) resource.TestCheckFunc {
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

func testAccLBTargetGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_lb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}

func testAccLBTargetGroupDynamicBlock(name string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "test-iig"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v2"
    description = "template_description"

    resources {
      memory = 2
      cores  = 2
    }

    boot_disk {
      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 2
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable = 3
    max_creating    = 3
    max_expansion   = 3
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "test-account"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_lb_target_group" "test-tg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  dynamic "target" {
     for_each = [for v in yandex_compute_instance_group.group1.instances.* : {
       subnet = v.network_interface.0.subnet_id,
       ip     = v.network_interface.0.ip_address
     }]

    content {
      subnet_id = target.value.subnet
      address   = target.value.ip
    }
  }
}
`, acceptance.FolderID(), name)
}
