package pluginsdk

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

var IamKMSSymmetricKeySchema = map[string]*schema.Schema{
	"symmetric_key_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Symmetric Key ID to apply a binding to.",
	},
}

type KMSSymmetricKeyIamUpdater struct {
	symmetricKeyID string
	Config         *providerconfig.Config
}

func newKMSSymmetricKeyIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &KMSSymmetricKeyIamUpdater{
		symmetricKeyID: d.Get("symmetric_key_id").(string),
		Config:         config,
	}, nil
}

func kmsSymmetricKeyIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("symmetric_key_id", d.Id())
	return nil
}

func (u *KMSSymmetricKeyIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetKMSSymmetricKeyAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *KMSSymmetricKeyIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	client := kmssdk.NewSymmetricKeyClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.symmetricKeyID,
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

func (u *KMSSymmetricKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	client := kmssdk.NewSymmetricKeyClient(u.Config.SDK)

	bSize := sdkcommon.YandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.symmetricKeyID,
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

func (u *KMSSymmetricKeyIamUpdater) GetResourceID() string {
	return u.symmetricKeyID
}

func (u *KMSSymmetricKeyIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-kms-symmetric-key-%s", u.symmetricKeyID)
}

func (u *KMSSymmetricKeyIamUpdater) DescribeResource() string {
	return fmt.Sprintf("KMS Symmetric Key '%s'", u.symmetricKeyID)
}

func GetKMSSymmetricKeyAccessBindings(ctx context.Context, config *providerconfig.Config, symmetricKeyID string) ([]*access.AccessBinding, error) {
	client := kmssdk.NewSymmetricKeyClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: symmetricKeyID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of %s: %w", symmetricKeyID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
