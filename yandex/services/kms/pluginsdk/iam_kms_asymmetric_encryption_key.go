package pluginsdk

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	kmsencryptionsdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricencryption"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

var IamKMSAsymmetricEncryptionKeySchema = map[string]*schema.Schema{
	"asymmetric_encryption_key_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Asymmetric Encryption Key ID to apply a binding to.",
	},
}

type KMSAsymmetricEncryptionKeyIamUpdater struct {
	asymmetricEncryptionKeyID string
	Config                    *providerconfig.Config
}

func newKMSAsymmetricEncryptionKeyIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &KMSAsymmetricEncryptionKeyIamUpdater{
		asymmetricEncryptionKeyID: d.Get("asymmetric_encryption_key_id").(string),
		Config:                    config,
	}, nil
}

func kmsAsymmetricEncryptionKeyIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("asymmetric_encryption_key_id", d.Id())
	return nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetKMSAsymmetricEncryptionKeyAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	client := kmsencryptionsdk.NewAsymmetricEncryptionKeyClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.asymmetricEncryptionKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, sdkcommon.YandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	client := kmsencryptionsdk.NewAsymmetricEncryptionKeyClient(u.Config.SDK)

	bSize := sdkcommon.YandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.asymmetricEncryptionKeyID,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) GetResourceID() string {
	return u.asymmetricEncryptionKeyID
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-kms-asymmetric-encryption-key-%s", u.asymmetricEncryptionKeyID)
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) DescribeResource() string {
	return fmt.Sprintf("KMS Asymmetric Encryption Key '%s'", u.asymmetricEncryptionKeyID)
}

func GetKMSAsymmetricEncryptionKeyAccessBindings(ctx context.Context, config *providerconfig.Config, asymmetricEncryptionKeyID string) ([]*access.AccessBinding, error) {
	client := kmsencryptionsdk.NewAsymmetricEncryptionKeyClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: asymmetricEncryptionKeyID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of %s: %w", asymmetricEncryptionKeyID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
