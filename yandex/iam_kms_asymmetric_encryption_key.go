package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamKMSAsymmetricEncryptionKeySchema = map[string]*schema.Schema{
	"asymmetric_encryption_key_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type KMSAsymmetricEncryptionKeyIamUpdater struct {
	asymmetricEncryptionKeyID string
	Config                    *Config
}

func newKMSAsymmetricEncryptionKeyIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &KMSAsymmetricEncryptionKeyIamUpdater{
		asymmetricEncryptionKeyID: d.Get("asymmetric_encryption_key_id").(string),
		Config:                    config,
	}, nil
}

func kmsAsymmetricEncryptionKeyIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("asymmetric_encryption_key_id", d.Id())
	return nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getKMSAsymmetricEncryptionKeyAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.asymmetricEncryptionKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *KMSAsymmetricEncryptionKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.asymmetricEncryptionKeyID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().UpdateAccessBindings(ctx, req))
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
				log.Printf("[DEBUG] request ID is %s\n", reqID)
			}
			return fmt.Errorf("Error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		err = op.Wait(ctx)
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

func getKMSAsymmetricEncryptionKeyAccessBindings(ctx context.Context, config *Config, asymmetricEncryptionKeyID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: asymmetricEncryptionKeyID,
			PageSize:   defaultListSize,
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
