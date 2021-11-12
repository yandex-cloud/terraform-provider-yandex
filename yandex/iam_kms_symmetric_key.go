package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMKMSDefaultTimeout = 1 * time.Minute

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

func (u *KMSSymmetricKeyIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getKMSSymmetricKeyAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *KMSSymmetricKeyIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.symmetricKeyID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMKMSDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.KMS().SymmetricKey().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
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

func getKMSSymmetricKeyAccessBindings(config *Config, symmetricKeyID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.KMS().SymmetricKey().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: symmetricKeyID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for KMS Symmetric Key %s: %s", symmetricKeyID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
