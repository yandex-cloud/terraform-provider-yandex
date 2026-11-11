package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	asymmetricencryptionsdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricencryption"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func registerYandexKmsAsymmetricEncryptionKeyResource(add func(string, *resource.Sweeper)) {
	add("yandex_kms_asymmetric_encryption_key", &resource.Sweeper{
		Name: "yandex_kms_asymmetric_encryption_key",
		F:    sweepAllKMSAsymmetricEncryptionKey,
	})
}

func sweepAllKMSAsymmetricEncryptionKey(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &kms.ListAsymmetricEncryptionKeysRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	resp, err := asymmetricencryptionsdk.NewAsymmetricEncryptionKeyClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting keys: %s", err)
	}
	result := &multierror.Error{}
	for _, k := range resp.Keys {
		id := k.GetId()
		if !test.SweepWithRetry(sweepKMSAsymmetricEncryptionKeyOnce, conf, "KMS Asymmetric Encryption Key", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep KSMS Asymmetric Encryption Key %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepKMSAsymmetricEncryptionKeyOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := asymmetricencryptionsdk.NewAsymmetricEncryptionKeyClient(conf.SDKv2).Update(ctx, &kms.UpdateAsymmetricEncryptionKeyRequest{
		KeyId:              id,
		DeletionProtection: false,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"deletion_protection"},
		},
	})
	if _, err = op.Wait(ctx); err != nil {
		return err
	}

	opDelete, err := asymmetricencryptionsdk.NewAsymmetricEncryptionKeyClient(conf.SDKv2).Delete(ctx, &kms.DeleteAsymmetricEncryptionKeyRequest{
		KeyId: id,
	})
	_, err = opDelete.Wait(ctx)
	return err
}
