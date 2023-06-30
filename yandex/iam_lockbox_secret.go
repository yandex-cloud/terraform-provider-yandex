package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"log"
	"time"
)

const yandexIAMLockboxDefaultTimeout = 1 * time.Minute
const yandexIAMLockboxUpdateAccessBindingsBatchSize = 1000

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

func (u *LockboxSecretIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getLockboxSecretAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *LockboxSecretIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.secretId,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMLockboxDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.LockboxSecret().Secret().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *LockboxSecretIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMLockboxUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.secretId,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.LockboxSecret().Secret().UpdateAccessBindings(ctx, req))
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

func (u *LockboxSecretIamUpdater) GetResourceID() string {
	return u.secretId
}

func (u *LockboxSecretIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-lockbox-secret-%s", u.secretId)
}

func (u *LockboxSecretIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Lockbox secret '%s'", u.secretId)
}

func getLockboxSecretAccessBindings(ctx context.Context, config *Config, secretId string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := config.sdk.LockboxSecret().Secret().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: secretId,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings of Lockbox secret %s: %w", secretId, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
