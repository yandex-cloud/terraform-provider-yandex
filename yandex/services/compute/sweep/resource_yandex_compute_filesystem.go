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

func registerResourceYandexComputeFilesystem(add func(string, *resource.Sweeper)) {
	add("yandex_compute_filesystem", &resource.Sweeper{
		Name: "yandex_compute_filesystem",
		F:    sweepAllComputeFilesystem,
		Dependencies: []string{
			"yandex_compute_instance",
		},
	})
}

func sweepAllComputeFilesystem(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListFilesystemsRequest{FolderId: conf.FolderID}
	it := computesdk.NewFilesystemClient(conf.SDK).Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		filesystem := it.Value()
		id := filesystem.GetId()
		if !sweepComputeFilesystem(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Filesystem %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeFilesystem(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepComputeFilesystemOnce, conf, "Compute Filesystem", id)
}

func sweepComputeFilesystemOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	op, err := computesdk.NewFilesystemClient(conf.
		SDK,
	).
		Delete(ctx, &compute.DeleteFilesystemRequest{
			FilesystemId: id,
		})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
