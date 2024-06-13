package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"golang.org/x/exp/slices"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

func init() {
	resource.AddTestSweepers("yandex_compute_instance_group", &resource.Sweeper{
		Name: "yandex_compute_instance_group",
		F:    testSweepComputeInstanceGroups,
		Dependencies: []string{
			"yandex_kubernetes_node_group",
		},
	})
}

func testSweepComputeInstanceGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	var serviceAccountID, networkID, subnetID string
	var depsCreated bool

	req := &instancegroup.ListInstanceGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.InstanceGroup().InstanceGroup().InstanceGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !depsCreated {
			depsCreated = true
			serviceAccountID, err = createIAMServiceAccountForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			networkID, err = createVPCNetworkForSweeper(conf)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
			subnetID, err = createVPCSubnetForSweeper(conf, networkID)
			if err != nil {
				result = multierror.Append(result, err)
				break
			}
		}

		id := it.Value().GetId()
		status := it.Value().GetStatus()
		if !updateComputeInstanceGroupWithSweeperDeps(conf, status, id, serviceAccountID, networkID, subnetID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep (update with dependencies) compute instance group %q", id))
			continue
		}

		if !sweepComputeInstanceGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute instance group %q", id))
		}
	}

	if serviceAccountID != "" {
		if !sweepIAMServiceAccount(conf, serviceAccountID) {
			result = multierror.Append(result,
				fmt.Errorf("failed to sweep IAM service account %q", serviceAccountID))
		}
	}
	if subnetID != "" {
		if !sweepVPCSubnet(conf, subnetID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC subnet %q", subnetID))
		}
	}
	if networkID != "" {
		if !sweepVPCNetwork(conf, networkID) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", networkID))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeInstanceGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeInstanceGroupOnce, conf, "Compute instance group", id)
}

func sweepComputeInstanceGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCNetworkDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.InstanceGroup().InstanceGroup().Delete(ctx, &instancegroup.DeleteInstanceGroupRequest{
		InstanceGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func updateComputeInstanceGroupWithSweeperDeps(conf *Config, status instancegroup.InstanceGroup_Status, instanceGroupID, serviceAccountID, networkID, subnetID string) bool {
	debugLog("started updating instance group %q", instanceGroupID)
	updateMaskPath := []string{
		"deletion_protection",
		"allocation_policy",
		"service_account_id",
		"instance_template.network_interface_specs",
	}
	if status == instancegroup.InstanceGroup_DELETING {
		updateMaskPath = []string{"service_account_id"}
	}

	client := conf.sdk.InstanceGroup().InstanceGroup()
	for i := 1; i <= conf.MaxRetries; i++ {
		req := &instancegroup.UpdateInstanceGroupRequest{
			InstanceGroupId:    instanceGroupID,
			DeletionProtection: false,
			ServiceAccountId:   serviceAccountID,
			AllocationPolicy: &instancegroup.AllocationPolicy{
				Zones: []*instancegroup.AllocationPolicy_Zone{
					{ZoneId: conf.Zone},
				},
			},
			InstanceTemplate: &instancegroup.InstanceTemplate{
				NetworkInterfaceSpecs: []*instancegroup.NetworkInterfaceSpec{
					{
						NetworkId:            networkID,
						SubnetIds:            []string{subnetID},
						PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{},
					},
				},
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: updateMaskPath,
			},
		}

		_, err := conf.sdk.WrapOperation(client.Update(conf.Context(), req))
		if err != nil {
			debugLog("[instance group %q] update try #%d: %v", instanceGroupID, i, err)
		} else {
			debugLog("[instance group %q] update try #%d: request was successfully sent", instanceGroupID, i)
			return true
		}
	}

	debugLog("instance group %q update failed", instanceGroupID)
	return false
}

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

func TestAccComputeInstanceGroup_Gpus(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigGpus(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupHasGpus(&ig, 1),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_NetworkSettings(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigNetworkSettings(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupNetworkSettings(&ig, "SOFTWARE_ACCELERATED"),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_MetadataOptions(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigMetadataOptions(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupMetadataOptions(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_Variables(t *testing.T) {
	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigVariables(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupVariables(&ig,
						append(make([]*instancegroup.Variable, 0),
							&instancegroup.Variable{Key: "test_key1", Value: "test_value1"},
							&instancegroup.Variable{Key: "test_key2", Value: "test_value2"})),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigVariables2(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
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
	t.Parallel()

	var ig instancegroup.InstanceGroup

	name := acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	sgName := acctest.RandomWithPrefix("tf-test")
	fsName1 := acctest.RandomWithPrefix("tf-test")
	fsName2 := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupConfigFull(name, saName, sgName, fsName1, fsName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDefaultValues(&ig),
					testAccCheckComputeInstanceGroupFixedScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_autoscale(t *testing.T) {
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
				Config: testAccComputeInstanceGroupConfigAutoScale(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupAutoScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_TestAutoScale(t *testing.T) {
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
				Config: testAccComputeInstanceGroupConfigTestAutoScale(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupTestAutoScalePolicy(&ig),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_DeployPolicyStrategy(t *testing.T) {
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
				Config: testAccComputeInstanceGroupConfigStrategy(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupStrategy(&ig),
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

func TestAccComputeInstanceGroup_DeletionProtection(t *testing.T) {
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
				Config: testAccComputeInstanceGroupConfigDeletionProtection(name, saName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDeletionProtection(&ig, true),
				),
			},
			{
				Config: testAccComputeInstanceGroupConfigDeletionProtection(name, saName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupDeletionProtection(&ig, false),
				),
			},
			computeInstanceGroupImportStep(),
		},
	})
}

func TestAccComputeInstanceGroup_createPlacementGroup(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckNonEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createAndErasePlacementGroup(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
			},
			{
				Config: testAccComputeInstanceGroupNoPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createAndChangePlacementGroup(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName1 := acctest.RandomWithPrefix("tf-test")
	pgName2 := acctest.RandomWithPrefix("tf-test")

	var pg1, pg2 string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccGetPlacementGroupIG(&ig, &pg1),
				),
			},
			{
				Config: testAccComputeInstanceGroupChangePlacementGroup(name, saName, pgName1, pgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccGetPlacementGroupIG(&ig, &pg2),
					testAccNotEqualStrings(&pg1, &pg2),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_createEmptyPlacementGroupAndAssignLater(t *testing.T) {
	t.Parallel()

	var ig instancegroup.InstanceGroup
	var name = acctest.RandomWithPrefix("tf-test")
	saName := acctest.RandomWithPrefix("tf-test")
	pgName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceGroupNoPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckEmptyPlacementGroupIG(&ig),
				),
			},
			{
				Config: testAccComputeInstanceGroupPlacementGroup(name, saName, pgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckNonEmptyPlacementGroupIG(&ig),
				),
			},
		},
	})
}

func TestAccComputeInstanceGroup_InstanceTagsPool(t *testing.T) {
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
				Config: testAccComputeInstanceGroupConfigInstanceTagsPool(name, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceGroupExists("yandex_compute_instance_group.group1", &ig),
					testAccCheckComputeInstanceGroupHasInstanceTagsPool(&ig),
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName, deletionProtection)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName, sgName, fsName1, fsName2)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName)
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
`, getExampleFolderID(), igName, saName, pgName)
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
`, getExampleFolderID(), igName, saName, pgName)
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
`, getExampleFolderID(), igName, saName, pgName1, pgName2)
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
		if ig.GetInstanceTemplate().ResourcesSpec.Memory != toBytes(2) {
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
