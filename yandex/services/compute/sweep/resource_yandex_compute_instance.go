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

func registerResourceYandexComputeInstance(add func(string, *resource.Sweeper)) {
	add("yandex_compute_instance", &resource.Sweeper{
		Name: "yandex_compute_instance",
		F:    sweepAllComputeInstances,
		Dependencies: []string{
			"yandex_dataproc_cluster",
			"yandex_kubernetes_cluster",
			"yandex_compute_instance_group",
		},
	})
}

func sweepAllComputeInstances(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListInstancesRequest{FolderId: conf.FolderID}
	it := computesdk.NewInstanceClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		instance := it.Value()
		id := instance.GetId()
		if !sweepComputeInstance(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Instance %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeInstance(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepComputeInstanceOnce, conf, "Compute Instance", id)
}

func sweepComputeInstanceOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_compute.YandexComputeInstanceDefaultTimeout)
	defer cancel()

	op, err := computesdk.NewInstanceClient(conf.
		SDK).
		Delete(ctx, &compute.DeleteInstanceRequest{
			InstanceId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
