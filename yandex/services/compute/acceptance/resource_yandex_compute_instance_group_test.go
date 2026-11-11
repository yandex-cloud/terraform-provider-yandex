package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/exp/slices"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	instancegroupsdk "github.com/yandex-cloud/go-sdk/services/compute/v1/instancegroup"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func computeInstanceGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_compute_instance_group.group1",
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccComputeInstanceGroup_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMain(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})

}

// TODO(fatalerr): merge into TestAccComputeInstanceGroup_full
// when Compute API supports it (CLOUD-244271).
func TestAccComputeInstanceGroup_reservedInstancePoolID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigReservedPool(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupReservedInstancePool(&ig, "test-reserved-pool"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_Gpus(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigGpus(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupHasGpus(&ig, 1),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_NetworkSettings(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigNetworkSettings(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupNetworkSettings(&ig, "SOFTWARE_ACCELERATED"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_MetadataOptions(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMetadataOptions(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupMetadataOptions(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_Variables(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigVariables(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupVariables(&ig,
						append(make([]*instancegroup.Variable, 0),
							&instancegroup.Variable{Key: "test_key1", Value: "test_value1"},
							&instancegroup.Variable{Key: "test_key2", Value: "test_value2"})),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigVariables2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupVariables(&ig,
						append(make([]*instancegroup.Variable, 0),
							&instancegroup.Variable{Key: "test_key1", Value: "test_value1_new"},
							&instancegroup.Variable{Key: "test_key2", Value: "test_value2"},
							&instancegroup.Variable{Key: "test_key3", Value: "test_value3"})),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	sgName := acctest.RandomWithPrefix("tf-test")
	fsName1 := acctest.RandomWithPrefix("tf-test")
	fsName2 := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigFull(name, saName, sgName, fsName1, fsName2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDefaultValues(&ig),
					testAccCheckComputeInstanceGroupFixedScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_autoscale(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigAutoScale(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupAutoScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_TestAutoScale(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigTestAutoScale(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTestAutoScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_DeployPolicyStrategy(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigStrategy(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupStrategy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithLabels(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithLabels2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key1", "label_value2"),
					testAccCheckComputeInstanceGroupLabel(&ig, "label_key_extra", "label_value_extra"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_update2(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels3(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTemplateLabel(&ig, "label_key1", "label_value1"),
					testAccCheckComputeInstanceGroupTemplateMeta(&ig, "meta_key1", "meta_val1"),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithTemplateLabels4(name, saName),

				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
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

func TestAccComputeInstanceGroup_updateFilesystem(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var fs1 compute.Filesystem
	var fs2 compute.Filesystem

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMain(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupNoFilesystemSpecs(&ig),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigWithFilesystemSpecs(name, saName, "fs1", "fs2"),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					suite.testAccCheckComputeFilesystemExists("yandex_compute_filesystem.inst-group-test-fs", &fs1),
					suite.testAccCheckComputeFilesystemExists("yandex_compute_filesystem.inst-group-test-fs2", &fs2),
					testAccCheckComputeInstanceGroupHasFilesystemSpecs(&ig, &fs1, &fs2),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_DeletionProtection(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigDeletionProtection(name, saName, true),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDeletionProtection(&ig, true),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigDeletionProtection(name, saName, false),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDeletionProtection(&ig, false),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_createPlacementGroup(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckNonEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createAndErasePlacementGroup(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
			},
			{
				Config: testAccComputeInstanceGroupNoPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createAndChangePlacementGroup(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName1 := acctest.RandomWithPrefix("tf-test")
	pgName2 := acctest.RandomWithPrefix("tf-test")

	var pg1, pg2 string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccGetPlacementGroupIG(&ig, &pg1),
				),
			},
			{
				Config: testAccComputeInstanceGroupChangePlacementGroup(name, saName, pgName1, pgName2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccGetPlacementGroupIG(&ig, &pg2),
					testAccNotEqualStrings(&pg1, &pg2),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createEmptyPlacementGroupAndAssignLater(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupNoPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckEmptyPlacementGroupIG(&ig),
				),
			},
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckNonEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_InstanceTagsPool(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigInstanceTagsPool(name, saName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupHasInstanceTagsPool(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func (suite testSuite) testAccCheckComputeInstanceGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance_group" {
			continue
		}

		_, err := instancegroupsdk.NewInstanceGroupClient(config.SDK).Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
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
	Name        string
}

type Filesystem struct {
	DeviceName string
	Mode       string
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
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigReservedPool(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v2"
    description = "template_description"

    reserved_instance_pool_id = "test-reserved-pool"

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
      size = 0 # Because test-reserved-pool does not exist.
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigWithFilesystemSpecs(igName string, saName string, fs1Name string, fs2Name string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_filesystem" "inst-group-test-fs" {
  name     = "%[4]s"
  size     = 10
  type     = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_filesystem" "inst-group-test-fs2" {
  name     = "%[5]s"
  size     = 15
  type     = "network-ssd"

  labels = {
    my-label = "my-label-value-2"
  }
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    filesystem {
      filesystem_id = "${yandex_compute_filesystem.inst-group-test-fs.id}"
      mode = "READ_WRITE"
    }

    filesystem {
      device_name = "fs2"
      filesystem_id = "${yandex_compute_filesystem.inst-group-test-fs2.id}"
      mode = "READ_WRITE"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, fs1Name, fs2Name)
}

func testAccComputeInstanceGroupConfigDeletionProtection(igName string, saName string, deletionProtection bool) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on          = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name                = "%[2]s"
  folder_id           = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id  = "${yandex_iam_service_account.test_account.id}"
  deletion_protection = "%[4]t"
  instance_template {
    platform_id = "standard-v2"
    description = "template_description"

    resources {
      memory        = 2
      cores         = 2
      core_fraction = 20
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
      size = 1
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, deletionProtection)
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
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
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
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

  labels = {
    label_key1      = "label_value2"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
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
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
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
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    labels = {
      label_key1      = "label_value2"
      label_key_extra = "label_value_extra"
    }

    metadata = {
      meta_key1      = "meta_val2"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigFull(igName, saName, sgName, fsName1, fsName2 string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "standard-v2"
    description = "template_description"
    name        = "my-instance-{instance.index}"
    hostname    = "my-hostname-{instance.index}"

    resources {
      memory        = 2
      cores         = 2
      core_fraction = 20
    }

    boot_disk {
      mode = "READ_WRITE"

      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
        type     = "network-hdd"
      }
    }

    secondary_disk {
      initialize_params {
        description = "desc1"
        image_id    = "${data.yandex_compute_image.ubuntu.id}"
        size        = 3
        type        = "network-ssd"
      }

      name = "secondary-disk-name1"
    }

    secondary_disk {
      initialize_params {
        description = "desc2"
        image_id    = "${data.yandex_compute_image.ubuntu.id}"
        size        = 3
        type        = "network-hdd"
      }

      name = "secondary-disk-name2"
    }

    filesystem {
      filesystem_id = "${yandex_compute_filesystem.inst-group-test-fs.id}"
      mode = "READ_WRITE"
    }

    filesystem {
      device_name = "fs2"
      filesystem_id = "${yandex_compute_filesystem.inst-group-test-fs2.id}"
      mode = "READ_WRITE"
    }

    network_interface {
      network_id         = "${yandex_vpc_network.inst-group-test-network.id}"
      subnet_ids         = ["${yandex_vpc_subnet.inst-group-test-subnet.id}"]
      security_group_ids = ["${yandex_vpc_security_group.sg1.id}"]
      dns_record {
        fqdn = "myhost.internal."
      }
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
    max_unavailable  = 4
    max_creating     = 3
    max_expansion    = 2
    max_deleting     = 1
    startup_duration = 5
  }
}

resource "yandex_compute_filesystem" "inst-group-test-fs" {
  name     = "%[5]s"
  size     = 10
  type     = "network-hdd"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_filesystem" "inst-group-test-fs2" {
  name     = "%[6]s"
  size     = 15
  type     = "network-ssd"

  labels = {
    my-label = "my-label-value-2"
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_security_group" "sg1" {
  depends_on  = ["yandex_vpc_network.inst-group-test-network"]
  name        = "%[4]s"
  description = "description"
  network_id  = "${yandex_vpc_network.inst-group-test-network.id}"
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"

  labels = {
    tf-label    = "tf-label-value-a"
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

resource "yandex_vpc_subnet" "inst-group-test-subnet" {
  depends_on     = ["yandex_vpc_network.inst-group-test-network"]
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, sgName, fsName1, fsName2)
}

func testAccComputeInstanceGroupConfigAutoScale(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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
        type     = "network-hdd"
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
    auto_scale {
      auto_scale_type = "REGIONAL"
      initial_size           = 1
      max_size               = 2
      min_zone_size          = 1
      measurement_duration   = 120
      cpu_utilization_target = 80
      custom_rule {
        rule_type   = "WORKLOAD"
        metric_type = "GAUGE"
        metric_name = "metric1"
        target      = 50
      }
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable  = 4
    max_creating     = 3
    max_expansion    = 2
    max_deleting     = 1
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigTestAutoScale(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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
      mode = "READ_WRITE"

      initialize_params {
        image_id = "${data.yandex_compute_image.ubuntu.id}"
        size     = 4
        type     = "network-hdd"
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
    test_auto_scale {
      auto_scale_type = "REGIONAL"
      initial_size           = 1
      max_size               = 2
      min_zone_size          = 1
      measurement_duration   = 120
      cpu_utilization_target = 80
      custom_rule {
        rule_type   = "WORKLOAD"
        metric_type = "GAUGE"
        metric_name = "metric1"
        target      = 50
      }
    }
  }

  allocation_policy {
    zones = ["ru-central1-a"]
  }

  deploy_policy {
    max_unavailable  = 4
    max_creating     = 3
    max_expansion    = 2
    max_deleting     = 1
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigGpus(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
  folder_id          = "${data.yandex_resourcemanager_folder.test_folder.id}"
  service_account_id = "${yandex_iam_service_account.test_account.id}"
  instance_template {
    platform_id = "gpu-standard-v2"
    description = "template_description"

    resources {
      cores  = 8
      memory = 48
      gpus   = 1
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
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
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
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigNetworkSettings(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    network_settings {
      type = "SOFTWARE_ACCELERATED"
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
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
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigMetadataOptions(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    network_settings {
      type = "SOFTWARE_ACCELERATED"
    }

	metadata_options {
	  gce_http_endpoint    = 1
	  aws_v1_http_endpoint = 1
	  gce_http_token       = 1
	  aws_v1_http_token    = 2
	  aws_v2_http_endpoint = 2
	  aws_v2_http_token    = 2
	}
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
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
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigVariables(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

  variables = {
    test_key1 = "test_value1"
    test_key2 = "test_value2"
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
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
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigVariables2(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

  variables = {
    test_key1 = "test_value1_new"
    test_key2 = "test_value2"
    test_key3 = "test_value3"
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    zones = ["ru-central1-b"]
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
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupConfigStrategy(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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
	strategy        = "opportunistic"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func testAccComputeInstanceGroupPlacementGroup(igName, saName, pgName string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    placement_policy {
      placement_group_id = yandex_compute_placement_group.pg.id
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

resource yandex_compute_placement_group pg {
  name        = "%[4]s"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, pgName)
}

func testAccComputeInstanceGroupNoPlacementGroup(igName, saName, pgName string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

resource yandex_compute_placement_group pg {
  name        = "%[4]s"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, pgName)
}

func testAccComputeInstanceGroupChangePlacementGroup(igName, saName, pgName1, pgName2 string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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

    placement_policy {
      placement_group_id = yandex_compute_placement_group.pg2.id
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

resource yandex_compute_placement_group pg {
  name        = "%[4]s"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
  }
}

resource yandex_compute_placement_group pg2 {
  name        = "%[5]s"
  description = "my description"
  labels = {
    first  = "xxx"
    second = "yyy"
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

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName, pgName1, pgName2)
}

func testAccComputeInstanceGroupConfigInstanceTagsPool(igName string, saName string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1604-lts"
}

data "yandex_resourcemanager_folder" "test_folder" {
  folder_id = "%[1]s"
}

resource "yandex_compute_instance_group" "group1" {
  depends_on         = ["yandex_iam_service_account.test_account", "yandex_resourcemanager_folder_iam_member.test_account"]
  name               = "%[2]s"
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
      subnet_ids = ["${yandex_vpc_subnet.subnet-a.id}", "${yandex_vpc_subnet.subnet-b.id}", "${yandex_vpc_subnet.subnet-c.id}"]
    }
  }

  scale_policy {
    fixed_scale {
      size = 3
    }
  }

  allocation_policy {
    zones = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
    instance_tags_pool {
      zone = "ru-central1-a"
      tags = ["atag"]
    }
    instance_tags_pool {
      zone = "ru-central1-b"
      tags = ["btag"]
    }
    instance_tags_pool {
      zone = "ru-central1-d"
      tags = ["ctag"]
    }
  }

  deploy_policy {
    max_unavailable = 1
    max_creating    = 3
    max_expansion   = 0
    max_deleting    = 3
  }
}

resource "yandex_vpc_network" "inst-group-test-network" {
  description = "tf-test"
}

resource "yandex_vpc_subnet" "subnet-a" {
  description    = "tf-test"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.11.0/24"]
}

resource "yandex_vpc_subnet" "subnet-b" {
  description    = "tf-test"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.12.0/24"]
}

resource "yandex_vpc_subnet" "subnet-c" {
  description    = "tf-test"
  zone           = "ru-central1-d"
  network_id     = "${yandex_vpc_network.inst-group-test-network.id}"
  v4_cidr_blocks = ["192.168.13.0/24"]
}

resource "yandex_iam_service_account" "test_account" {
  name        = "%[3]s"
  description = "tf-test"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "${data.yandex_resourcemanager_folder.test_folder.id}"
  member      = "serviceAccount:${yandex_iam_service_account.test_account.id}"
  role        = "editor"
  sleep_after = 30
}
`, acceptance.FolderID(), igName, saName)
}

func (suite testSuite) testAccCheckComputeInstanceGroupExists(n string, instance *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		found, err := instancegroupsdk.NewInstanceGroupClient(config.SDK).Get(context.Background(), &instancegroup.GetInstanceGroupRequest{
			InstanceGroupId: rs.Primary.ID,
			View:            instancegroup.InstanceGroupView_FULL,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("instancegroup is not found")
		}

		//goland:noinspection GoVetCopyLock
		*instance = *found

		return nil
	}
}

func testAccCheckComputeInstanceGroupHasGpus(ig *instancegroup.InstanceGroup, gpus int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.GetInstanceTemplate().ResourcesSpec.Gpus != gpus {
			return fmt.Errorf("invalid resources.gpus value in instance_template in instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupVariables(ig *instancegroup.InstanceGroup, variables []*instancegroup.Variable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(ig.GetVariables()) != len(variables) {
			return fmt.Errorf("invalid variables value in instance group %s", ig.Name)
		}
		for _, raw1 := range variables {
			var a = false
			for _, raw2 := range ig.GetVariables() {
				if raw1.GetKey() == raw2.GetKey() && raw1.GetValue() == raw2.GetValue() {
					a = true
				}
			}
			if !a {
				return fmt.Errorf("invalid variables value in instance group %s", ig.Name)
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupNetworkSettings(ig *instancegroup.InstanceGroup, nst string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.GetInstanceTemplate().GetNetworkSettings().GetType().String() != nst {
			return fmt.Errorf("invalid network_settings.type value in instance_template in instance group %s", ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.Labels == nil {
			return fmt.Errorf("no labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupDeletionProtection(ig *instancegroup.InstanceGroup, value bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.DeletionProtection != value {
			return fmt.Errorf("expected value '%t' but found value '%t' for deletion_protection field on instance group %s", value, !value, ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateLabel(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetLabels() == nil {
			return fmt.Errorf("no template labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupTemplateMeta(ig *instancegroup.InstanceGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.GetMetadata() == nil {
			return fmt.Errorf("no template labels found on instance group %s", ig.Name)
		}

		if v, ok := ig.InstanceTemplate.Metadata[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on instance group %s template labels", value, v, key, ig.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on instance group %s template labels", key, ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupNoFilesystemSpecs(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(ig.InstanceTemplate.FilesystemSpecs) != 0 {
			return fmt.Errorf("invalid number of filesystem specs in instance group %s", ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupHasFilesystemSpecs(ig *instancegroup.InstanceGroup, fs1 *compute.Filesystem, fs2 *compute.Filesystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(ig.InstanceTemplate.FilesystemSpecs) != 2 {
			return fmt.Errorf("invalid number of filesystem specs in instance group %s", ig.Name)
		}
		attachedFs1 := &Filesystem{Mode: "READ_WRITE"}
		attachedFs2 := &Filesystem{DeviceName: "fs2", Mode: "READ_WRITE"}
		for _, spec := range ig.InstanceTemplate.FilesystemSpecs {
			var err error
			if spec.FilesystemId == fs1.Id {
				err = checkFs(fs1.Name, spec, attachedFs1)
			} else if spec.FilesystemId == fs2.Id {
				err = checkFs(fs2.Name, spec, attachedFs2)
			} else {
				err = fmt.Errorf("unknown filesystem spec with id %s in instance group %s", spec.FilesystemId, ig.Name)
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupDefaultValues(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// InstanceTemplate
		if ig.GetInstanceTemplate() == nil {
			return fmt.Errorf("no InstanceTemplate in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().PlatformId != "standard-v2" {
			return fmt.Errorf("invalid PlatformId value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().Description != "template_description" {
			return fmt.Errorf("invalid Description value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().Name != "my-instance-{instance.index}" {
			return fmt.Errorf("invalid name value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().Hostname != "my-hostname-{instance.index}" {
			return fmt.Errorf("invalid hostname value in instance group %s", ig.Name)
		}
		// Resources
		if ig.GetInstanceTemplate().ResourcesSpec == nil {
			return fmt.Errorf("no ResourcesSpec in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Cores != 2 {
			return fmt.Errorf("invalid Cores value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.Memory != sdkcommon.ToBytes(2) {
			return fmt.Errorf("invalid Memory value in instance group %s", ig.Name)
		}
		if ig.GetInstanceTemplate().ResourcesSpec.CoreFraction != 20 {
			return fmt.Errorf("invalid CoreFraction value in instance group %s", ig.Name)
		}
		// SchedulingPolicy
		if !ig.GetInstanceTemplate().SchedulingPolicy.Preemptible {
			return fmt.Errorf("invalid Preemptible value in instance group %s", ig.Name)
		}
		// BootDisk
		bootDisk := &Disk{Mode: "READ_WRITE", Size: 4, Type: "network-hdd"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s boot disk", ig.Name), ig.GetInstanceTemplate().BootDiskSpec, bootDisk); err != nil {
			return err
		}
		// SecondaryDisk
		if len(ig.InstanceTemplate.SecondaryDiskSpecs) != 2 {
			return fmt.Errorf("invalid number of secondary disks in instance group %s", ig.Name)
		}

		disk0 := &Disk{Size: 3, Type: "network-ssd", Description: "desc1", Name: "secondary-disk-name1"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #0", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[0], disk0); err != nil {
			return err
		}

		disk1 := &Disk{Size: 3, Type: "network-hdd", Description: "desc2", Name: "secondary-disk-name2"}
		if err := checkDisk(fmt.Sprintf("instancegroup %s secondary disk #1", ig.Name), ig.InstanceTemplate.SecondaryDiskSpecs[1], disk1); err != nil {
			return err
		}

		if len(ig.InstanceTemplate.FilesystemSpecs) != 2 {
			return fmt.Errorf("invalid number of filesystems in instance group %s", ig.Name)
		}

		fs0 := &Filesystem{Mode: "READ_WRITE"}
		fs1 := &Filesystem{DeviceName: "fs2", Mode: "READ_WRITE"}
		for _, spec := range ig.InstanceTemplate.FilesystemSpecs {
			err1 := checkFs(fmt.Sprintf("instancegroup %s attached file system", ig.Name), spec, fs0)
			err2 := checkFs(fmt.Sprintf("instancegroup %s attached file system", ig.Name), spec, fs1)

			if err1 != nil && err2 != nil {
				return err1
			}
		}

		// NetworkInterfaceSpec
		if len(ig.GetInstanceTemplate().GetNetworkInterfaceSpecs()) != 1 {
			return fmt.Errorf("expected 1 network_interface_spec, got %d", len(ig.GetInstanceTemplate().GetNetworkInterfaceSpecs()))
		}

		// NetworkSettings
		if ig.GetInstanceTemplate().GetNetworkInterfaceSpecs()[0].SecurityGroupIds == nil || len(ig.GetInstanceTemplate().GetNetworkInterfaceSpecs()[0].SecurityGroupIds) == 0 {
			return fmt.Errorf("invalid network_interface.security_group_ids value in instance group %s", ig.Name)
		}

		// AllocationPolicy
		if ig.AllocationPolicy == nil || len(ig.AllocationPolicy.Zones) != 1 || ig.AllocationPolicy.Zones[0].ZoneId != "ru-central1-a" {
			return fmt.Errorf("invalid allocation policy in instance group %s", ig.Name)
		}

		// Deploy policy
		if ig.GetDeployPolicy() == nil {
			return fmt.Errorf("no deploy policy in instance group %s", ig.Name)
		}

		if ig.GetDeployPolicy().MaxUnavailable != 4 {
			return fmt.Errorf("invalid MaxUnavailable in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxCreating != 3 {
			return fmt.Errorf("invalid MaxCreating in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxExpansion != 2 {
			return fmt.Errorf("invalid MaxExpansion in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().MaxDeleting != 1 {
			return fmt.Errorf("invalid MaxDeleting in instance group %s", ig.Name)
		}
		if ig.GetDeployPolicy().StartupDuration.Seconds != 5 {
			return fmt.Errorf("invalid StartupDuration in instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupFixedScalePolicy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.ScalePolicy.GetFixedScale() == nil || ig.ScalePolicy.GetFixedScale().Size != 2 {
			return fmt.Errorf("invalid fixed scale policy on instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupAutoScalePolicy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.ScalePolicy.GetAutoScale() == nil {
			return fmt.Errorf("no auto scale policy on instance group %s", ig.Name)
		}

		sp := ig.ScalePolicy.GetAutoScale()
		if sp.AutoScaleType != instancegroup.ScalePolicy_AutoScale_REGIONAL {
			return fmt.Errorf("wrong auto_scale_type on instance group %s", ig.Name)
		}
		if sp.InitialSize != 1 {
			return fmt.Errorf("wrong initialsize on instance group %s", ig.Name)
		}
		if sp.MaxSize != 2 {
			return fmt.Errorf("wrong max_size on instance group %s", ig.Name)
		}
		if sp.MeasurementDuration == nil || sp.MeasurementDuration.Seconds != 120 {
			return fmt.Errorf("wrong measurement_duration on instance group %s", ig.Name)
		}
		if sp.CpuUtilizationRule == nil || sp.CpuUtilizationRule.UtilizationTarget != 80. {
			return fmt.Errorf("wrong cpu_utilization_target on instance group %s", ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupTestAutoScalePolicy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.ScalePolicy.GetTestAutoScale() == nil {
			return fmt.Errorf("no test auto scale policy on instance group %s", ig.Name)
		}

		sp := ig.ScalePolicy.GetTestAutoScale()
		if sp.AutoScaleType != instancegroup.ScalePolicy_AutoScale_REGIONAL {
			return fmt.Errorf("wrong auto_scale_type on instance group %s", ig.Name)
		}
		if sp.InitialSize != 1 {
			return fmt.Errorf("wrong initial size on instance group %s", ig.Name)
		}
		if sp.MaxSize != 2 {
			return fmt.Errorf("wrong max_size on instance group %s", ig.Name)
		}
		if sp.MeasurementDuration == nil || sp.MeasurementDuration.Seconds != 120 {
			return fmt.Errorf("wrong measurement_duration on instance group %s", ig.Name)
		}
		if sp.CpuUtilizationRule == nil || sp.CpuUtilizationRule.UtilizationTarget != 80. {
			return fmt.Errorf("wrong cpu_utilization_target on instance group %s", ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupStrategy(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.DeployPolicy == nil {
			return fmt.Errorf("no deploy policy on instance group %s", ig.Name)
		}

		if ig.DeployPolicy.Strategy.String() != "OPPORTUNISTIC" {
			return fmt.Errorf("wrong deploy_policy.strategy on instance group %s", ig.Name)
		}

		return nil
	}
}

func testAccCheckComputeInstanceGroupMetadataOptions(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.InstanceTemplate.MetadataOptions == nil {
			return fmt.Errorf("no metadata options on instance group %s", ig.Name)
		}

		if ig.InstanceTemplate.MetadataOptions.GceHttpEndpoint.String() != "ENABLED" {
			return fmt.Errorf("wrong metadata_options.gce_http_endpoint on instance group %s", ig.Name)
		}
		if ig.InstanceTemplate.MetadataOptions.AwsV1HttpEndpoint.String() != "ENABLED" {
			return fmt.Errorf("wrong metadata_options.aws_v1_http_endpoint on instance group %s", ig.Name)
		}
		if ig.InstanceTemplate.MetadataOptions.GceHttpToken.String() != "ENABLED" {
			return fmt.Errorf("wrong metadata_options.gce_http_token on instance group %s", ig.Name)
		}
		if ig.InstanceTemplate.MetadataOptions.AwsV1HttpToken.String() != "DISABLED" {
			return fmt.Errorf("wrong metadata_options.aws_v1_http_token on instance group %s", ig.Name)
		}
		if ig.InstanceTemplate.MetadataOptions.AwsV2HttpEndpoint.String() != "DISABLED" {
			return fmt.Errorf("wrong metadata_options.aws_v2_http_endpoint on instance group %s", ig.Name)
		}
		if ig.InstanceTemplate.MetadataOptions.AwsV2HttpToken.String() != "DISABLED" {
			return fmt.Errorf("wrong metadata_options.aws_v2_http_token on instance group %s", ig.Name)
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
	if a.DiskSpec.Size != sdkcommon.ToBytes(d.Size) {
		return fmt.Errorf("invalid Size value in %s", name)
	}
	if a.DiskSpec.GetSnapshotId() != d.Snapshot {
		return fmt.Errorf("invalid Snapshot value in %s", name)
	}
	if d.Image != "" && a.DiskSpec.GetImageId() != d.Image {
		return fmt.Errorf("invalid Image value in %s", name)
	}
	if a.Name != d.Name {
		return fmt.Errorf("invalid Name value in %s", name)
	}
	return nil
}

func checkFs(name string, spec *instancegroup.AttachedFilesystemSpec, f *Filesystem) error {
	if f.Mode != "" && spec.Mode.String() != f.Mode {
		return fmt.Errorf("invalid Mode value in %s", name)
	}
	if f.DeviceName != "" && spec.DeviceName != f.DeviceName {
		return fmt.Errorf("invalid DeviceName value in %s", name)
	}
	return nil
}

func testAccCheckNonEmptyPlacementGroupIG(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if ig.InstanceTemplate.PlacementPolicy != nil && ig.InstanceTemplate.PlacementPolicy.PlacementGroupId != "" {
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccCheckEmptyPlacementGroupIG(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if ig.InstanceTemplate.PlacementPolicy == nil || ig.InstanceTemplate.PlacementPolicy.PlacementGroupId == "" {
			return nil
		}
		return fmt.Errorf("instance placement_group_id is not empty (%s), ig %s",
			ig.InstanceTemplate.PlacementPolicy.PlacementGroupId, ig.Id)
	}
}

func testAccGetPlacementGroupIG(ig *instancegroup.InstanceGroup, pg *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if ig.InstanceTemplate.PlacementPolicy != nil {
			*pg = ig.InstanceTemplate.PlacementPolicy.PlacementGroupId
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccCheckComputeInstanceGroupReservedInstancePool(ig *instancegroup.InstanceGroup, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if ig.GetInstanceTemplate().ReservedInstancePoolId != expected {
			return fmt.Errorf("invalid reserved instance pool id value %q in instance group %s", ig.GetInstanceTemplate().ReservedInstancePoolId, ig.Name)
		}
		return nil
	}
}

func testAccCheckComputeInstanceGroupHasInstanceTagsPool(ig *instancegroup.InstanceGroup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if ig.AllocationPolicy == nil || len(ig.AllocationPolicy.Zones) != 3 {
			return fmt.Errorf("invalid allocation policy in instance group %s", ig.Name)
		}
		err := assertZoneExists(ig.AllocationPolicy.Zones, "ru-central1-a", []string{"atag"})
		if err != nil {
			return err
		}
		err = assertZoneExists(ig.AllocationPolicy.Zones, "ru-central1-b", []string{"btag"})
		if err != nil {
			return err
		}
		err = assertZoneExists(ig.AllocationPolicy.Zones, "ru-central1-d", []string{"ctag"})
		if err != nil {
			return err
		}
		return nil
	}
}

func assertZoneExists(zones []*instancegroup.AllocationPolicy_Zone, zoneId string, tags []string) error {
	found := false
	for _, zone := range zones {
		if zone.ZoneId == zoneId && slices.Equal(tags, zone.InstanceTagsPool) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Allocation zone not found.\nExpected zoneId = %s, tags = %+v\nGot zones: %+v", zoneId, tags, zones)
	}
	return nil
}
