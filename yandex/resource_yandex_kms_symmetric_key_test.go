package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"
	"google.golang.org/grpc/codes"
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
			"yandex_mdb_opensearch_cluster",
		},
	})
}

func testSweepKMSSymmetricKey(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	client := kmssdk.NewSymmetricKeyClient(conf.SDK)

	req := &kms.ListSymmetricKeysRequest{FolderId: conf.FolderID}
	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepKMSSymmetricKey(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSM symmetric key %q", id))
		}
	}
	if err := it.Error(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepKMSSymmetricKey(conf *Config, id string) bool {
	return sweepWithRetry(sweepKMSSymmetricKeyOnce, conf, "KMS Symmetric Key", id)
}

func sweepKMSSymmetricKeyOnce(conf *Config, id string) error {
	client := kmssdk.NewSymmetricKeyClient(conf.SDK)

	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := client.Delete(ctx, &kms.DeleteSymmetricKeyRequest{
		KeyId: id,
	})
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
