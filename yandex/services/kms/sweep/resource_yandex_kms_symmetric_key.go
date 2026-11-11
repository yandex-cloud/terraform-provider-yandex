package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc/codes"
)

func registerResourceYandexKmsSymmetricKey(add func(string, *resource.Sweeper)) {
	add("yandex_kms_symmetric_key", &resource.Sweeper{
		Name: "yandex_kms_symmetric_key",
		F:    sweepAllKMSSymmetricKey,
		Dependencies: []string{
			"yandex_compute_instance",
			"yandex_compute_instance_group",
			"yandex_compute_disk",
			"yandex_kubernetes_cluster",
			"yandex_mdb_opensearch_cluster",
		},
	})
}

func sweepAllKMSSymmetricKey(_ string) error {
	conf, err := sweep.ConfigForSweepers()
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

func sweepKMSSymmetricKey(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepKMSSymmetricKeyOnce, conf, "KMS Symmetric Key", id)
}

func sweepKMSSymmetricKeyOnce(conf *providerconfig.Config, id string) error {
	client := kmssdk.NewSymmetricKeyClient(conf.SDK)

	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := client.Delete(ctx, &kms.DeleteSymmetricKeyRequest{
		KeyId: id,
	})
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
