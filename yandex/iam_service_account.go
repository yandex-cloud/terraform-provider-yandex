package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamServiceAccountSchema = map[string]*schema.Schema{
	"service_account_id": {
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
	bindings, err := getServiceAccountAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ServiceAccountIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.serviceAccountID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMServiceAccountDefaultTimeout)
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

func getServiceAccountAccessBindings(config *Config, serviceAccountID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.IAM().ServiceAccount().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: serviceAccountID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for service account %s: %s", serviceAccountID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
