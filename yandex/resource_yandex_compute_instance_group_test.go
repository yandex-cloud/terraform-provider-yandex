package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

func computeInstanceGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_compute_instance_group.group1",
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccComputeInstanceGroup_basic(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMain(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_full(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigFull(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDefaultValues(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithLabels(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithLabels2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value2"),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key_extra", "label_value_extra"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update2(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels3(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key1", "label_value1"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key1", "meta_val1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels4(name, saName),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key1", "label_value2"),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key_extra", "label_value_extra"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key1", "meta_val2"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key_extra", "meta_value_extra"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func testAccCheckComputeInstanceGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance_group" {
			continue
		}

		_, err := config.sdk.InstanceGroup().InstanceGroup().Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
			InstanceGroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Instance Group still exists")
		}
	}

	return nil
}

type Disk struct {
	Description string
	Mode        string
	Size        int
	Type        string
	Image       string
	Snapshot    string
}

func testAccComputeInstanceGroupConfigMain(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithLabels(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
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

  labels = {
    label_key1 = "label_value1"
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithLabels2(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
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

  labels = {
    label_key1 = "label_value2"
    label_key_extra = "label_value_extra"
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithTemplateLabels3(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
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

    labels = {
      label_key1 = "label_value1"
    }

    metadata = {
      meta_key1 = "meta_val1"
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithTemplateLabels4(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
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

    labels = {
      label_key1 = "label_value2"
      label_key_extra = "label_value_extra"
    }

    metadata = {
      meta_key1 = "meta_val2"
      meta_key_extra = "meta_value_extra"
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigFull(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_binding.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v1"
    description = "template_description"

    resources {
      memory = 2
      cores  = 1
      core_fraction = 20
    }

    boot_disk {
      mode = "READ_ONLY"

      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
        type = "network-nvme"
      }
    }

    secondary_disk {
      initialize_params {
        description = "desc1"
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 3
        type = "network-nvme"
      }
    }

    secondary_disk {
      initialize_params {
        description = "desc2"
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 3
        type = "network-hdd"
      }
    }

    network_interface {
      network_id = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
    }

    scheduling_policy {
        preemptible = true
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
    max_unavailable = 4
    max_creating    = 3
    max_expansion   = 2
    max_deleting    = 1
    startup_duration = 5
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_binding" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  members     = ["serviceAccount:${yandex_iam_service_account.test_account.id}"]
  role        = "editor"
  sleep_after = 30
}
`, getExampleFolderID(), igName, saName)

}

func testAccCheckComputeInstanceGroupExists(n string, instance *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.InstanceGroup().InstanceGroup().Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
			InstanceGroupId: rs.Primary.ID,
			View:            instancegroup.InstanceGroupView_FULL,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("instancegroup is not found")
		}

		*instance = *found

		return nil
	}
}

func testAccCheckComputeInstanceGroupLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.Labels == nil {
			return fmt.Errorf("no labels found on instancegroup %s", ig.Name)
		}

		if v, ok := ig.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instancegroup %s", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instancegroup %s", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetLabels() == nil {
			return fmt.Errorf("no template labels found on instancegroup %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instancegroup %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instancegroup %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateMeta(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetMetadata() == nil {
			return fmt.Errorf("no template labels found on instancegroup %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Metadata[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instancegroup %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instancegroup %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupDefaultValues(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// ScalePolicy
		if ig.ScalePolicy.GetFixedScale() == nil || ig.ScalePolicy.GetFixedScale().Size != 2 {
			return fmt.Errorf("invalid scale policy on instancegroup %s", ig.Name)
		}
		// InstanceTemplate
		if ig.GetInstanceTemplate() == nil {
			return fmt.Errorf("no InstanceTemplate in instancegroup %s", ig.Name)
		}
		if ig.GetInstanceTemplate().PlatformId != "standard-v1" {
			return fmt.Errorf("invalid PlatformId value in instancegroup %s", ig.Name)
		}
		if ig.GetInstanceTemplate().Description != "template_description" {
			return fmt.Errorf("invalid Description value in instancegroup %s", ig.Name)
		}
		// Resources
		if ig.GetInstanceTemplate().ResourcesSpec == nil {
			return fmt.Errorf("no ResourcesSpec in instancegroup %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Cores != 1 {
			return fmt.Errorf("invalid Cores value in instancegroup %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Memory != toBytes(2) {
			return fmt.Errorf("invalid Memory value in instancegroup %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.CoreFraction != 20 {
			return fmt.Errorf("invalid CoreFraction value in instancegroup %s", ig.Name)
		}
		// SchedulingPolicy
		if !ig.GetInstanceTemplate().SchedulingPolicy.Preemptible {
			return fmt.Errorf("invalid Preemptible value in instancegroup %s", ig.Name)
		}
		// BootDisk
		bootDisk := &Disk{Mode: "READ_ONLY", Size: 4, Type: "network-nvme"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s boot disk", ig.Name), ig.GetInstanceTemplate().BootDiskSpec, bootDisk); err != nil {
			return err
		}
		// SecondaryDisk
		if len(ig.InstanceTemplate.SecondaryDiskSpecs) != 2 {
			return fmt.Errorf("invalid number of secondary disks in instancegroup %s", ig.Name)
		}

		disk0 := &Disk{Size: 3, Type: "network-nvme", Description: "desc1"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #0", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[0], disk0); err != nil {
			return err
		}

		disk1 := &Disk{Size: 3, Type: "network-hdd", Description: "desc2"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #1", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[1], disk1); err != nil {
			return err
		}

		// AllocationPolicy
		if ig.AllocationPolicy == nil || len(ig.AllocationPolicy.Zones) != 1 || ig.AllocationPolicy.Zones[0].ZoneId != "ru-central1-a" {
			return fmt.Errorf("invalid allocation policy in instancegroup %s", ig.Name)
		}

		// Deploy policy
		if ig.GetDeployPolicy() == nil {
			return fmt.Errorf("no deploy policy in instancegroup %s", ig.Name)
		}

		if ig.GetDeployPolicy().MaxUnavailable != 4 {
			return fmt.Errorf("invalid MaxUnavailable in instancegroup %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxCreating != 3 {
			return fmt.Errorf("invalid MaxCreating in instancegroup %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxExpansion != 2 {
			return fmt.Errorf("invalid MaxExpansion in instancegroup %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxDeleting != 1 {
			return fmt.Errorf("invalid MaxDeleting in instancegroup %s", ig.Name)
		}
		if ig.GetDeployPolicy().StartupDuration.Seconds != 5 {
			return fmt.Errorf("invalid StartupDuration in instancegroup %s", ig.Name)
		}

		return nil
	}
}

func checkDisk(name string, a *instancegroup.AttachedDiskSpec, d *Disk) error {
	if d.Mode != "" && a.Mode.String() != d.Mode {
		return fmt.Errorf("invalid Mode value in %s", name)
	}
	if a.DiskSpec.Description != d.Description {
		return fmt.Errorf("invalid Description value in %s", name)
	}
	if d.Type != "" && a.DiskSpec.TypeId != d.Type {
		return fmt.Errorf("invalid Type value in %s", name)
	}
	if a.DiskSpec.Size != toBytes(d.Size) {
		return fmt.Errorf("invalid Size value in %s", name)
	}
	if a.DiskSpec.GetSnapshotId() != d.Snapshot {
		return fmt.Errorf("invalid Snapshot value in %s", name)
	}
	if d.Image != "" && a.DiskSpec.GetImageId() != d.Image {
		return fmt.Errorf("invalid Image value in %s", name)
	}
	return nil
}
