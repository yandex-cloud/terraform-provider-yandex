package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMKMSDefaultTimeout = 1 * time.Minute
const yandexIAMKMSUpdateAccessBindingsBatchSize = 1000

var IamKMSSymmetricKeySchema = map[string]*schema.Schema{
	"symmetric_key_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type KMSSymmetricKeyIamUpdater struct {
	symmetricKeyID string
	Config         *Config
}

func newKMSSymmetricKeyIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &KMSSymmetricKeyIamUpdater{
		symmetricKeyID: d.Get("symmetric_key_id").(string),
		Config:         config,
	}, nil
}

func kmsSymmetricKeyIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("symmetric_key_id", d.Id())
	return nil
}

func (u *KMSSymmetricKeyIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getKMSSymmetricKeyAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *KMSSymmetricKeyIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.symmetricKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMS().SymmetricKey().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *KMSSymmetricKeyIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMKMSUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.symmetricKeyID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMS().SymmetricKey().UpdateAccessBindings(ctx, req))
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

func (u *KMSSymmetricKeyIamUpdater) GetResourceID() string {
	return u.symmetricKeyID
}

func (u *KMSSymmetricKeyIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-kms-symmetric-key-%s", u.symmetricKeyID)
}

func (u *KMSSymmetricKeyIamUpdater) DescribeResource() string {
	return fmt.Sprintf("KMS Symmetric Key '%s'", u.symmetricKeyID)
}

func getKMSSymmetricKeyAccessBindings(ctx context.Context, config *Config, symmetricKeyID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.KMS().SymmetricKey().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: symmetricKeyID,
			PageSize:   defaultListSize,
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
