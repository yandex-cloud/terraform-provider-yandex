package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_compute "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/pluginsdk"
)

func registerResourceYandexComputeDisk(add func(string, *resource.Sweeper)) {
	add("yandex_compute_disk", &resource.Sweeper{
		Name: "yandex_compute_disk",
		F:    sweepAllComputeDisks,
		Dependencies: []string{
			"yandex_compute_instance",
			"yandex_compute_instance_group",
		},
	})
}

func sweepAllComputeDisks(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListDisksRequest{FolderId: conf.FolderID}
	it := computesdk.NewDiskClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		disk := it.Value()
		id := disk.GetId()
		if !sweepComputeDisk(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Disk %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeDisk(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepComputeDiskOnce, conf, "Compute Disk", id)
}

func sweepComputeDiskOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_compute.YandexComputeDiskDefaultTimeout)
	defer cancel()

	op, err := computesdk.NewDiskClient(conf.
		SDK).
		Delete(ctx, &compute.DeleteDiskRequest{
			DiskId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
