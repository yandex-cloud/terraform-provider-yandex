package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

func init() {
	resource.AddTestSweepers("yandex_kms_symmetric_key", &resource.Sweeper{
		Name: "yandex_kms_symmetric_key",
		F:    testSweepKMSSymmetricKey,
		Dependencies: []string{
			"yandex_compute_instance",
			"yandex_compute_instance_group",
			"yandex_compute_disk",
			"yandex_kubernetes_cluster",
		},
	})
}

func testSweepKMSSymmetricKey(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &kms.ListSymmetricKeysRequest{FolderId: conf.FolderID}
	it := conf.sdk.KMS().SymmetricKey().SymmetricKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKMSSymmetricKey(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSM symmetric key %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKMSSymmetricKey(conf *Config, id string) bool {
	return sweepWithRetry(sweepKMSSymmetricKeyOnce, conf, "KMS Symmetric Key", id)
}

func sweepKMSSymmetricKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := conf.sdk.KMS().SymmetricKey().Delete(ctx, &kms.DeleteSymmetricKeyRequest{
		KeyId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
