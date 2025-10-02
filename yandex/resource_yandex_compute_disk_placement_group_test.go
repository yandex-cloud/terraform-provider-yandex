package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_disk_placement_group", &resource.Sweeper{
		Name: "yandex_compute_disk_placement_group",
		F:    testSweepComputeDiskPlacementGroups,
		Dependencies: []string{
			"yandex_compute_disk",
		},
	})
}

func sweepComputeDiskPlacementGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := conf.sdk.Compute().DiskPlacementGroup().Delete(ctx, &compute.DeleteDiskPlacementGroupRequest{
		DiskPlacementGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepComputeDiskPlacementGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListDiskPlacementGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().DiskPlacementGroup().DiskPlacementGroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(sweepComputeDiskPlacementGroupOnce, conf, "Placement group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute Placement Group %q", id))
		}
	}

	return result.ErrorOrNil()
}
