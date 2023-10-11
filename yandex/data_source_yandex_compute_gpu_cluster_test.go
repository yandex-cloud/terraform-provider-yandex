package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceComputeGpuCluster_byID(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeGpuClusterDestroy,
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeGpuClusterDestroy,
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
