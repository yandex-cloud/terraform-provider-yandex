package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	op, err := conf.sdk.Compute().GpuCluster().Delete(ctx, &compute.DeleteGpuClusterRequest{
		GpuClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
