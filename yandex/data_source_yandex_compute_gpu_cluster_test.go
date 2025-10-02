package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func TestAccDataSourceComputeGpuCluster_byID(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomGpuClusterConfig(gpuClusterName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_gpu_cluster.source", "gpu_cluster_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"name", gpuClusterName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"description"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"interconnect_type", "infiniband"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"zone"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"folder_id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"status"),
					testAccCheckCreatedAtAttr("data.yandex_compute_gpu_cluster.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeGpuCluster_byName(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomGpuClusterConfig(gpuClusterName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_compute_gpu_cluster.source", "gpu_cluster_id"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"name", gpuClusterName),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"description"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("data.yandex_compute_gpu_cluster.source",
						"interconnect_type", "infiniband"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"zone"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"folder_id"),
					resource.TestCheckResourceAttrSet("data.yandex_compute_gpu_cluster.source",
						"status"),
					testAccCheckCreatedAtAttr("data.yandex_compute_gpu_cluster.source"),
				),
			},
		},
	})
}

func testAccCheckComputeGpuClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_gpu_cluster" {
			continue
		}

		_, err := config.sdk.Compute().GpuCluster().Get(context.Background(), &compute.GetGpuClusterRequest{
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

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.Compute().GpuCluster().Get(context.Background(), &compute.GetGpuClusterRequest{
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

func testAccDataSourceCustomGpuClusterResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foo" {
  name              = "%s"
  description       = "GPU cluster description"
  zone              = "ru-central1-a"
  interconnect_type = "infiniband"

  labels = {
    my-label = "my-label-value"
  }
}
`, name)
}

func testAccDataSourceCustomGpuClusterConfig(name string, useID bool) string {
	if useID {
		return testAccDataSourceCustomGpuClusterResourceConfig(name) + computeGpuClusterDataByIDConfig
	}

	return testAccDataSourceCustomGpuClusterResourceConfig(name) + computeGpuClusterDataByNameConfig
}

const computeGpuClusterDataByIDConfig = `
data "yandex_compute_gpu_cluster" "source" {
  gpu_cluster_id = "${yandex_compute_gpu_cluster.foo.id}"
}
`

const computeGpuClusterDataByNameConfig = `
data "yandex_compute_gpu_cluster" "source" {
  name = "${yandex_compute_gpu_cluster.foo.name}"
}
`
