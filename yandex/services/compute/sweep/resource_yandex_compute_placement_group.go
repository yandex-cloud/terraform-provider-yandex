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

func registerResourceYandexComputePlacementGroup(add func(string, *resource.Sweeper)) {
	add("yandex_compute_placement_group", &resource.Sweeper{
		Name: "yandex_compute_placement_group",
		F:    sweepAllComputePlacementGroups,
		Dependencies: []string{
			"yandex_compute_instance_group",
			"yandex_compute_instance",
		},
	})
}

func sweepAllComputePlacementGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListPlacementGroupsRequest{FolderId: conf.FolderID}
	it := computesdk.NewPlacementGroupClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		group := it.Value()
		id := group.GetId()
		if !sweep.SweepWithRetry(sweepComputePlacementGroupOnce, conf, "Placement group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep compute Placement Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputePlacementGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_compute.YandexComputePlacementGroupDefaultTimeout)
	defer cancel()

	op, err := computesdk.NewPlacementGroupClient(conf.
		SDK,
	).
		Delete(ctx, &compute.DeletePlacementGroupRequest{
			PlacementGroupId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
