package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceComputeGpuCluster_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomGpuClusterConfig(gpuClusterName, true),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_gpu_cluster.source", "gpu_cluster_id"),
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
					statecheck.CheckCreatedAtAttr("data.yandex_compute_gpu_cluster.source"),
				),
			},
		},
	})
}

func TestAccDataSourceComputeGpuCluster_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCustomGpuClusterConfig(gpuClusterName, false),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_compute_gpu_cluster.source", "gpu_cluster_id"),
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
					statecheck.CheckCreatedAtAttr("data.yandex_compute_gpu_cluster.source"),
				),
			},
		},
	})
}

func (suite testSuite) testAccCheckComputeGpuClusterDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_gpu_cluster" {
			continue
		}

		_, err := computesdk.NewGpuClusterClient(config.SDK).Get(context.Background(), &compute.GetGpuClusterRequest{
			GpuClusterId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("GPU cluster still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckComputeGpuClusterExists(n string, gpuCluster *compute.GpuCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		found, err := computesdk.NewGpuClusterClient(config.SDK).Get(context.Background(), &compute.GetGpuClusterRequest{
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
  interconnect_type = "INFINIBAND"

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
