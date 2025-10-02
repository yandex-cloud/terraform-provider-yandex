package yandex_compute_gpu_cluster_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccComputeGpuCluster_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccComputeGpuCluster_old(gpuClusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar", "name", gpuClusterName),
					resource.TestCheckResourceAttrSet("yandex_compute_gpu_cluster.foobar", "zone"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"interconnect_type", "infiniband"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccComputeGpuCluster_basic(gpuClusterName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccComputeGpuCluster_basic(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")
	var gpuCluster compute.GpuCluster

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeGpuCluster_basic(gpuClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeGpuClusterExists("yandex_compute_gpu_cluster.foobar", &gpuCluster),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar", "name", gpuClusterName),
					resource.TestCheckResourceAttrSet("yandex_compute_gpu_cluster.foobar", "zone"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"interconnect_type", "INFINIBAND"),
				),
			},
		},
	})
}

func TestAccComputeGpuCluster_update(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")
	var gpuCluster compute.GpuCluster

	newgpuClusterName := acctest.RandomWithPrefix("tf-test")
	newDesc := "new description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckComputeGpuClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeGpuCluster_basic(gpuClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeGpuClusterExists("yandex_compute_gpu_cluster.foobar", &gpuCluster),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar", "name", gpuClusterName),
					resource.TestCheckResourceAttrSet("yandex_compute_gpu_cluster.foobar", "zone"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"labels.my-label", "my-label-value"),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar",
						"interconnect_type", "INFINIBAND"),
				),
			},
			{
				Config: testAccComputeGpuCluster_updated(newgpuClusterName, newDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeGpuClusterExists("yandex_compute_gpu_cluster.foobar", &gpuCluster),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar", "name", newgpuClusterName),
					resource.TestCheckResourceAttr("yandex_compute_gpu_cluster.foobar", "description", newDesc),
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

		_, err := computesdk.NewGpuClusterClient(config.SDKv2).Get(context.Background(), &compute.GetGpuClusterRequest{
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

		found, err := computesdk.NewGpuClusterClient(config.SDKv2).Get(context.Background(), &compute.GetGpuClusterRequest{
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
func testAccComputeGpuCluster_old(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foobar" {
  name              = "%s"
  interconnect_type = "infiniband"

  labels = {
    my-label = "my-label-value"
  }
}
`, name)
}

func testAccComputeGpuCluster_basic(name string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foobar" {
  name              = "%s"
  interconnect_type = "INFINIBAND"

  labels = {
    my-label = "my-label-value"
  }
}
`, name)
}

//revive:disable:var-naming
func testAccComputeGpuCluster_updated(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foobar" {
  name              = "%s"
  description       = "%s"
  interconnect_type = "INFINIBAND"

  labels = {
    my-label = "my-label-value"
  }
}
`, name, desc)
}
