package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	kmssignaturesdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricsignature"
)

var IamKMSAsymmetricSignatureKeySchema = map[string]*schema.Schema{
	"asymmetric_signature_key_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The [Yandex Key Management Service](https://yandex.cloud/docs/kms/) Asymmetric Signature Key ID to apply a binding to.",
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
	client := kmssignaturesdk.NewAsymmetricSignatureKeyClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.asymmetricSignatureKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMKMSDefaultTimeout)
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

func (u *KMSAsymmetricSignatureKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	client := kmssignaturesdk.NewAsymmetricSignatureKeyClient(u.Config.SDK)

	bSize := yandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.asymmetricSignatureKeyID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := isRequestIDPresent(err); ok {
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
	client := kmssignaturesdk.NewAsymmetricSignatureKeyClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
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
