package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexComputeGpuCluster(add func(string, *resource.Sweeper)) {
	add("yandex_compute_gpu_cluster", &resource.Sweeper{
		Name: "yandex_compute_gpu_cluster",
		F:    sweepAllComputeGpuCluster,
		Dependencies: []string{
			"yandex_compute_instance",
		},
	})
}

func sweepAllComputeGpuCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListGpuClustersRequest{FolderId: conf.FolderID}
	it := computesdk.NewGpuClusterClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		cluster := it.Value()
		id := cluster.GetId()
		if !sweepComputeGpuCluster(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute GPU Cluster %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeGpuCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepComputeGpuClusterOnce, conf, "Compute GPU Cluster", id)
}

func sweepComputeGpuClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	op, err := computesdk.NewGpuClusterClient(conf.
		SDK,
	).
		Delete(ctx, &compute.DeleteGpuClusterRequest{
			GpuClusterId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
