package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_placement_group", &resource.Sweeper{
		Name: "yandex_compute_placement_group",
		F:    testSweepComputePlacementGroups,
		Dependencies: []string{
			"yandex_compute_instance_group",
			"yandex_compute_instance",
		},
	})
}

func sweepComputePlacementGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexComputePlacementGroupDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Compute().PlacementGroup().Delete(ctx, &compute.DeletePlacementGroupRequest{
		PlacementGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepComputePlacementGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListPlacementGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().PlacementGroup().PlacementGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(sweepComputePlacementGroupOnce, conf, "Placement group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute Placement Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestAccComputeInstance_createPlacementGroup(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstancePlacementGroupWithPartitionStrategy(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckNonEmptyPlacementGroup(&instance),
				),
			},
		},
	})
}

func TestAccComputeInstance_createAndErasePlacementGroup(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstancePlacementGroup(instanceName),
			},
			{
				Config: testAccComputeInstanceNoPlacementGroup(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckEmptyPlacementGroup(&instance),
				),
			},
		},
	})
}

func TestAccComputeInstance_createAndChangePlacementGroup(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))
	var pg1, pg2 string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstancePlacementGroup(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccGetPlacementGroup(&instance, &pg1),
				),
			},
			{
				Config: testAccComputeInstanceChangePlacementGroup(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccGetPlacementGroup(&instance, &pg2),
					testAccNotEqualStrings(&pg1, &pg2),
				),
			},
		},
	})
}

func TestAccComputeInstance_createEmptyPlacementGroupAndAssignLater(t *testing.T) {
	t.Parallel()

	var instance compute.Instance
	var instanceName = fmt.Sprintf("instance-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstanceNoPlacementGroup(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckEmptyPlacementGroup(&instance),
				),
			},
			{
				Config: testAccComputeInstancePlacementGroup(instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists("yandex_compute_instance.foobar", &instance),
					testAccCheckNonEmptyPlacementGroup(&instance),
				),
			},
		},
	})
}

func testAccCheckNonEmptyPlacementGroup(instance *compute.Instance) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if instance.PlacementPolicy != nil && instance.PlacementPolicy.PlacementGroupId != "" {
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccCheckEmptyPlacementGroup(instance *compute.Instance) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if instance.PlacementPolicy != nil && instance.PlacementPolicy.PlacementGroupId == "" {
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccGetPlacementGroup(instance *compute.Instance, pg *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if instance.PlacementPolicy != nil {
			*pg = instance.PlacementPolicy.PlacementGroupId
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccNotEqualStrings(s1 *string, s2 *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if *s1 != *s2 {
			return nil
		}
		return fmt.Errorf("instance placement_group_id is invalid")
	}
}

func testAccComputeInstancePlacementGroup(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
  placement_policy {
    placement_group_id = yandex_compute_placement_group.pg.id
  }
}

resource yandex_compute_placement_group pg {
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstancePlacementGroupWithPartitionStrategy(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
  placement_policy {
    placement_group_id = yandex_compute_placement_group.pg.id
    placement_group_partition = 3
  }
}

resource yandex_compute_placement_group pg {
	placement_strategy_partitions = 3
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstanceNoPlacementGroup(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
  placement_policy {
    placement_group_id = ""
  }
}

resource yandex_compute_placement_group pg {
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}

func testAccComputeInstanceChangePlacementGroup(instance string) string {
	// language=tf
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"
  allow_stopping_for_update = true

  resources {
    cores  = 2
    memory = 2
  }

  boot_disk {
    initialize_params {
      size     = 4
      image_id = "${data.yandex_compute_image.ubuntu.id}"
    }
  }

  network_interface {
    subnet_id = "${yandex_vpc_subnet.inst-test-subnet.id}"
  }

  metadata = {
    foo = "bar"
    baz = "qux"
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
  placement_policy {
    placement_group_id = yandex_compute_placement_group.pg2.id
  }
}

resource yandex_compute_placement_group pg {
}
resource yandex_compute_placement_group pg2 {
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}
`, instance)
}
