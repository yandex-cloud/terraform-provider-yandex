package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"time"
)

const yandexIAMLockboxDefaultTimeout = 1 * time.Minute

var IamLockboxSecretSchema = map[string]*schema.Schema{
	"secret_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type LockboxSecretIamUpdater struct {
	secretId string
	Config   *Config
}

func newLockboxSecretIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &LockboxSecretIamUpdater{
		secretId: d.Get("secret_id").(string),
		Config:   config,
	}, nil
}

func LockboxSecretIDParseFunc(d *schema.ResourceData, _ *Config) error {
	err := d.Set("secret_id", d.Id())
	if err != nil {
		return err
	}
	return nil
}

func (u *LockboxSecretIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getLockboxSecretAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *LockboxSecretIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.secretId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMLockboxDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.LockboxSecret().Secret().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil
}

func (u *LockboxSecretIamUpdater) GetResourceID() string {
	return u.secretId
}

func (u *LockboxSecretIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-lockbox-secret-%s", u.secretId)
}

func (u *LockboxSecretIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Lockbox secret '%s'", u.secretId)
}

func getLockboxSecretAccessBindings(config *Config, secretId string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.LockboxSecret().Secret().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: secretId,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("error retrieving IAM access bindings for Lockbox secret %s: %s", secretId, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
