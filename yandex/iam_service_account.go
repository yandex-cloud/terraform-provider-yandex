package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamServiceAccountSchema = map[string]*schema.Schema{
	"service_account_id": &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type ServiceAccountIamUpdater struct {
	serviceAccountID string
	Config           *Config
}

func newServiceAccountIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &ServiceAccountIamUpdater{
		serviceAccountID: d.Get("service_account_id").(string),
		Config:           config,
	}, nil
}

func serviceAccountIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("service_account_id", d.Id())
	return nil
}

func (u *ServiceAccountIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	resp, err := u.Config.sdk.IAM().ServiceAccount().ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
		ResourceId: u.serviceAccountID,
	})

	if err != nil {
		return nil, fmt.Errorf("Error retrieving IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return &Policy{resp.AccessBindings}, nil
}

func (u *ServiceAccountIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.serviceAccountID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(context.Background(), yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.IAM().ServiceAccount().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil

}

func (u *ServiceAccountIamUpdater) GetResourceID() string {
	return u.serviceAccountID
}

func (u *ServiceAccountIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-service-account-%s", u.serviceAccountID)
}

func (u *ServiceAccountIamUpdater) DescribeResource() string {
	return fmt.Sprintf("service account '%s'", u.serviceAccountID)
}
