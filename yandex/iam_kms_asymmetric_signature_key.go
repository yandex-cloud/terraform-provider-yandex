package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamKMSAsymmetricSignatureKeySchema = map[string]*schema.Schema{
	"asymmetric_signature_key_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type KMSAsymmetricSignatureKeyIamUpdater struct {
	asymmetricSignatureKeyID string
	Config                   *Config
}

func newKMSAsymmetricSignatureKeyIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &KMSAsymmetricSignatureKeyIamUpdater{
		asymmetricSignatureKeyID: d.Get("asymmetric_signature_key_id").(string),
		Config:                   config,
	}, nil
}

func kmsAsymmetricSignatureKeyIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("asymmetric_signature_key_id", d.Id())
	return nil
}

func (u *KMSAsymmetricSignatureKeyIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getKMSAsymmetricSignatureKeyAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *KMSAsymmetricSignatureKeyIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.asymmetricSignatureKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *KMSAsymmetricSignatureKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.asymmetricSignatureKeyID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().UpdateAccessBindings(ctx, req))
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

func (u *KMSAsymmetricSignatureKeyIamUpdater) GetResourceID() string {
	return u.asymmetricSignatureKeyID
}

func (u *KMSAsymmetricSignatureKeyIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-kms-asymmetric-signature-key-%s", u.asymmetricSignatureKeyID)
}

func (u *KMSAsymmetricSignatureKeyIamUpdater) DescribeResource() string {
	return fmt.Sprintf("KMS Asymmetric Signature Key '%s'", u.asymmetricSignatureKeyID)
}

func getKMSAsymmetricSignatureKeyAccessBindings(ctx context.Context, config *Config, asymmetricSignatureKeyID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: asymmetricSignatureKeyID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of %s: %w", asymmetricSignatureKeyID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
