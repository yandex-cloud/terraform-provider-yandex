package compute_instance_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	instanceResource = "yandex_compute_instance.foobar"
	timeout          = 15 * time.Minute
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeInstance_basic1IamMember(t *testing.T) {
	var (
		instance     compute.Instance
		instanceName = fmt.Sprintf("%s-%s", test.TestPrefix(), fmt.Sprintf("instance-test-%s", acctest.RandString(10)))
		userID       = "allUsers"
		role         = "editor"
		ctx, cancel  = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeInstance_basic(instanceName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeInstanceExists(instanceResource, &instance),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().Instance()
					}, &instance, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

//revive:disable:var-naming
func testAccComputeInstance_basic(instance, role, userID string) string {
	return fmt.Sprintf(`
data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-1804-lts"
}

resource "yandex_compute_instance" "foobar" {
  name        = "%s"
  description = "testAccComputeInstance_basic"
  platform_id = "standard-v2"
  zone        = "ru-central1-a"

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

  metadata_options {
    gce_http_endpoint = 1
    aws_v1_http_endpoint = 1
    gce_http_token = 1
    aws_v1_http_token = 2
  }

  labels = {
    my_key       = "my_value"
    my_other_key = "my_other_value"
  }
}

resource "yandex_vpc_network" "inst-test-network" {}

resource "yandex_vpc_subnet" "inst-test-subnet" {
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.inst-test-network.id}"
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_compute_instance_iam_binding" "test-compute-instance-bind" {
    role = "%s"
    members = ["system:%s"]
    instance_id = yandex_compute_instance.foobar.id
}
`, instance, role, userID)
}

func testAccCheckComputeInstanceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_instance" {
			continue
		}

		_, err := config.SDK.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Instance still exists")
		}
	}

	return nil
}

func testAccCheckComputeInstanceExists(n string, instance *compute.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().Instance().Get(context.Background(), &compute.GetInstanceRequest{
			InstanceId: rs.Primary.ID,
			View:       compute.InstanceView_FULL,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}

		*instance = *found

		return nil
	}
}
