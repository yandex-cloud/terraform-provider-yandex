package yandex_compute_reserved_instance_pool_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeReservedInstancePool_basic(t *testing.T) {
	t.Parallel()

	instancePoolName := acctest.RandomWithPrefix("tf-instance-pool")
	cfg, model := testAccReservedInstancePool_basic(instancePoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckReservedInstancePoolDestroy),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReservedInstancePoolExists("yandex_compute_reserved_instance_pool.pool"),
					testAccCheckReservedInstancePoolEqual("yandex_compute_reserved_instance_pool.pool", model),
				),
			},
		},
	})
}

func TestAccComputeReservedInstancePool_update(t *testing.T) {
	t.Parallel()

	instancePoolName := acctest.RandomWithPrefix("tf-instance-pool")
	cfg, model := testAccReservedInstancePool_basic(instancePoolName)
	updateCfg, updateModel := testAccReservedInstancePool_update("updated-reserved-instance-pool")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckReservedInstancePoolDestroy),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReservedInstancePoolExists("yandex_compute_reserved_instance_pool.pool"),
					testAccCheckReservedInstancePoolEqual("yandex_compute_reserved_instance_pool.pool", model),
				),
			},
			{
				Config: updateCfg,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if proto.Equal(model, updateModel) {
							return fmt.Errorf(
								"model should not be equal: before: %s, after: %s",
								ProtoToDebugString(model),
								ProtoToDebugString(updateModel),
							)
						}

						return nil
					},
					testAccCheckReservedInstancePoolExists("yandex_compute_reserved_instance_pool.pool"),
					testAccCheckReservedInstancePoolEqual("yandex_compute_reserved_instance_pool.pool", updateModel),
					func(s *terraform.State) error {
						if model.Id != updateModel.Id {
							return fmt.Errorf(
								"pool was unexpected recreated: got %s, expected %s",
								updateModel.Id,
								model.Id,
							)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccComputeReservedInstancePool_recreate(t *testing.T) {
	t.Parallel()

	instancePoolName := acctest.RandomWithPrefix("tf-instance-pool")
	cfg, model := testAccReservedInstancePool_basic(instancePoolName)
	updateCfg, updateModel := testAccReservedInstancePool_recreate(instancePoolName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckReservedInstancePoolDestroy),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReservedInstancePoolExists("yandex_compute_reserved_instance_pool.pool"),
					testAccCheckReservedInstancePoolEqual("yandex_compute_reserved_instance_pool.pool", model),
				),
			},
			{
				Config: updateCfg,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						if proto.Equal(model, updateModel) {
							return fmt.Errorf(
								"model should not be equal: before: %s, after: %s",
								ProtoToDebugString(model),
								ProtoToDebugString(updateModel),
							)
						}

						return nil
					},
					testAccCheckReservedInstancePoolExists("yandex_compute_reserved_instance_pool.pool"),
					testAccCheckReservedInstancePoolEqual("yandex_compute_reserved_instance_pool.pool", updateModel),
					func(s *terraform.State) error {
						if model.Id == updateModel.Id {
							return fmt.Errorf(
								"pool was not recreated: got %s, expected %s",
								updateModel.Id,
								model.Id,
							)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccCheckReservedInstancePoolDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_reserved_instance_pool" {
			continue
		}

		_, err := computesdk.NewReservedInstancePoolClient(config.SDKv2).Get(context.Background(), &compute.GetReservedInstancePoolRequest{
			ReservedInstancePoolId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Reserved Instance Pool still exists")
		}
	}

	return nil
}

func testAccCheckReservedInstancePoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := computesdk.NewReservedInstancePoolClient(config.SDKv2).Get(context.Background(), &compute.GetReservedInstancePoolRequest{
			ReservedInstancePoolId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("ReservedInstancePool %s not found", n)
		}

		return nil
	}
}

func testAccCheckReservedInstancePoolEqual(n string, reservedInstancePool *compute.ReservedInstancePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := computesdk.NewReservedInstancePoolClient(config.SDKv2).Get(context.Background(), &compute.GetReservedInstancePoolRequest{
			ReservedInstancePoolId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}
		// only for test - model fields depend on resources or cloud/folder info
		reservedInstancePool.Id = found.Id
		clearDynamicFields(found)

		if !proto.Equal(reservedInstancePool, found) {
			return fmt.Errorf(
				"Found difference between model and config: got: %s, found: %s",
				ProtoToDebugString(reservedInstancePool),
				ProtoToDebugString(found),
			)
		}

		return nil
	}
}

func ProtoToDebugString(msg proto.Message) string {
	if msg == nil {
		return "<nil>"
	}

	opts := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: true,
	}

	data, err := opts.Marshal(msg)
	if err != nil {
		return fmt.Sprintf("<error marshaling to json: %v>", err)
	}

	return string(data)
}

func testAccReservedInstancePool_basic(poolName string) (string, *compute.ReservedInstancePool) {
	cfg := fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_reserved_instance_pool" "pool" {
  name        = "%s"
  description = "reserved-instance-pool"

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  zone        = "ru-central1-a"

  # folder_id
  platform_id = "standard-v2"

  resources_spec = {
    cores         = 2
    core_fraction = 100
    memory        = 2147483648
  }

  # gpu settings

  boot_disk_spec = {
	image_id = "${data.yandex_compute_image.ubuntu.id}"
  }

  network_settings = {
    type = "STANDARD"
  }

  size = 1
  # allow_oversubscription
}
`, poolName)

	model := &compute.ReservedInstancePool{
		Name:        poolName,
		Description: "reserved-instance-pool",
		Labels: map[string]string{
			"my_key":       "my_value",
			"my_other_key": "my_other_value",
		},
		ZoneId:     "ru-central1-a",
		PlatformId: "standard-v2",
		ResourcesSpec: &compute.ResourcesSpec{
			Cores:        2,
			CoreFraction: 100,
			Memory:       2147483648,
		},
		GpuSettings: &compute.GpuSettings{
			GpuClusterId: "",
		},
		NetworkSettings: &compute.NetworkSettings{
			Type: compute.NetworkSettings_STANDARD,
		},
		Size: 1,
	}

	return cfg, model
}

func testAccReservedInstancePool_update(poolName string) (string, *compute.ReservedInstancePool) {
	cfg := fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_reserved_instance_pool" "pool" {
  name        = "%s"
  description = "new-description-for-reserved-instance-pool"

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
	another_value = "another_value"
  }

  zone        = "ru-central1-a"

  # folder_id
  platform_id = "standard-v2"

  resources_spec = {
    cores         = 2
    core_fraction = 100
    memory        = 2147483648
  }

  # gpu settings

  boot_disk_spec = {
	image_id = "${data.yandex_compute_image.ubuntu.id}"
  }

  network_settings = {
    type = "STANDARD"
  }

  size = 2
  allow_oversubscription = true
}
`, poolName)

	model := &compute.ReservedInstancePool{
		Name:        poolName,
		Description: "new-description-for-reserved-instance-pool",
		Labels: map[string]string{
			"my_key":        "my_value",
			"my_other_key":  "my_other_value",
			"another_value": "another_value",
		},
		ZoneId:     "ru-central1-a",
		PlatformId: "standard-v2",
		ResourcesSpec: &compute.ResourcesSpec{
			Cores:        2,
			CoreFraction: 100,
			Memory:       2147483648,
		},
		GpuSettings: &compute.GpuSettings{
			GpuClusterId: "",
		},
		NetworkSettings: &compute.NetworkSettings{
			Type: compute.NetworkSettings_STANDARD,
		},
		Size:                  2,
		AllowOversubscription: true,
	}

	return cfg, model
}

func testAccReservedInstancePool_recreate(poolName string) (string, *compute.ReservedInstancePool) {
	cfg := fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_reserved_instance_pool" "pool" {
  name        = "%s"
  description = "reserved-instance-pool"

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }

  zone        = "ru-central1-a"

  # folder_id
  platform_id = "standard-v2"

  resources_spec = {
    cores         = 4
    core_fraction = 100
    memory        = 4294967296
  }

  # gpu settings

  boot_disk_spec = {
	image_id = "${data.yandex_compute_image.ubuntu.id}"
  }

  network_settings = {
    type = "STANDARD"
  }

  size = 1
  allow_oversubscription = true
}
`, poolName)

	model := &compute.ReservedInstancePool{
		Name:        poolName,
		Description: "reserved-instance-pool",
		Labels: map[string]string{
			"my_key":       "my_value",
			"my_other_key": "my_other_value",
		},
		ZoneId:     "ru-central1-a",
		PlatformId: "standard-v2",
		ResourcesSpec: &compute.ResourcesSpec{
			Cores:        4,
			CoreFraction: 100,
			Memory:       4294967296,
		},
		GpuSettings: &compute.GpuSettings{
			GpuClusterId: "",
		},
		NetworkSettings: &compute.NetworkSettings{
			Type: compute.NetworkSettings_STANDARD,
		},
		Size:                  1,
		AllowOversubscription: true,
	}

	return cfg, model
}

func clearDynamicFields(model *compute.ReservedInstancePool) {
	model.FolderId = ""
	model.CloudId = ""
	model.CreatedAt = nil
	model.ProductIds = nil
	model.CommittedSize = 0
	model.SlotStats = nil
	model.InstanceStats = nil
}
