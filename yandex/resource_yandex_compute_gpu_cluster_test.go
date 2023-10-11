package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_gpu_cluster", &resource.Sweeper{
		Name: "yandex_compute_gpu_cluster",
		F:    testSweepComputeGpuCluster,
		Dependencies: []string{
			"yandex_compute_instance",
		},
	})
}

func testSweepComputeGpuCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListGpuClustersRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().GpuCluster().GpuClusterIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepComputeGpuCluster(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute GPU Cluster %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeGpuCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeGpuClusterOnce, conf, "Compute GPU Cluster", id)
}

func sweepComputeGpuClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexComputeGpuClusterDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Compute().GpuCluster().Delete(ctx, &compute.DeleteGpuClusterRequest{
		GpuClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccComputeGpuCluster_basic(t *testing.T) {
	t.Parallel()

	gpuClusterName := acctest.RandomWithPrefix("tf-test")
	var gpuCluster compute.GpuCluster

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckComputeGpuClusterDestroy,
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
						"interconnect_type", "infiniband"),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckComputeGpuClusterDestroy,
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
						"interconnect_type", "infiniband"),
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

//revive:disable:var-naming
func testAccComputeGpuCluster_basic(name string) string {
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

//revive:disable:var-naming
func testAccComputeGpuCluster_updated(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_compute_gpu_cluster" "foobar" {
  name              = "%s"
  description       = "%s"
  interconnect_type = "infiniband"

  labels = {
    my-label = "my-label-value"
  }
}
`, name, desc)
}
