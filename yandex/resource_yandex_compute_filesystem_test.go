package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func init() {
	resource.AddTestSweepers("yandex_compute_filesystem", &resource.Sweeper{
		Name: "yandex_compute_filesystem",
		F:    testSweepComputeFilesystem,
		Dependencies: []string{
			"yandex_compute_instance",
		},
	})
}

func testSweepComputeFilesystem(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &compute.ListFilesystemsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Compute().Filesystem().FilesystemIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepComputeFilesystem(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Compute Filesystem %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepComputeFilesystem(conf *Config, id string) bool {
	return sweepWithRetry(sweepComputeFilesystemOnce, conf, "Compute Filesystem", id)
}

func sweepComputeFilesystemOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(5 * time.Minute)
	defer cancel()

	op, err := conf.sdk.Compute().Filesystem().Delete(ctx, &compute.DeleteFilesystemRequest{
		FilesystemId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
