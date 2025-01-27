package compute_gpu_cluster_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const timeout = 15 * time.Minute

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeGpuCluster_basicIamMember(t *testing.T) {
	var (
		gpuClusterName = test.GenerateNameForResource(10)
		gpuCluster     compute.GpuCluster
		userID         = "allUsers"
		role           = "editor"
		ctx, cancel    = context.WithTimeout(context.Background(), timeout)
	)
	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeGpuCluster_basic(gpuClusterName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeGpuClusterExists("yandex_compute_gpu_cluster.foobar", &gpuCluster),
					test.TestAccCheckIamBindingExists(ctx, func() test.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return cfg.SDK.Compute().GpuCluster()
					}, &gpuCluster, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckComputeGpuClusterDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_gpu_cluster" {
			continue
		}

		_, err := config.SDK.Compute().GpuCluster().Get(context.Background(), &compute.GetGpuClusterRequest{
			GpuClusterId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("GPU cluster still exists")
		}
	}

	return nil
}

func testAccCheckComputeGpuClusterExists(n string, gpuCluster *compute.GpuCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.Compute().GpuCluster().Get(context.Background(), &compute.GetGpuClusterRequest{
			GpuClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("GPU cluster not found")
		}

		*gpuCluster = *found

		return nil
	}
}

//revive:disable:var-naming
func testAccComputeGpuCluster_basic(name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foobar" {
  name              = "%s"
  interconnect_type = "infiniband"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_compute_gpu_cluster_iam_binding" "test-gpu-binding" {
    role = "%s"
    members = ["system:%s"]
    gpu_cluster_id = yandex_compute_gpu_cluster.foobar.id
}
`, name, role, userID)
}
