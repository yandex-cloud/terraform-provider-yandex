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

func registerResourceYandexComputeDiskPlacementGroup(add func(string, *resource.Sweeper)) {
	add("yandex_compute_disk_placement_group", &resource.Sweeper{
		Name: "yandex_compute_disk_placement_group",
		F:    sweepAllComputeDiskPlacementGroups,
		Dependencies: []string{
			"yandex_compute_disk",
		},
	})
}

func sweepAllComputeDiskPlacementGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListDiskPlacementGroupsRequest{FolderId: conf.FolderID}
	it := computesdk.NewDiskPlacementGroupClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		group := it.Value()
		id := group.GetId()
		if !sweep.SweepWithRetry(sweepComputeDiskPlacementGroupOnce, conf, "Placement group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute Placement Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeDiskPlacementGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := computesdk.NewDiskPlacementGroupClient(conf.
		SDK).
		Delete(ctx, &compute.DeleteDiskPlacementGroupRequest{
			DiskPlacementGroupId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
