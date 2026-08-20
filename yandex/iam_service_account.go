package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
)

const yandexIAMServiceAccountUpdateAccessBindingsBatchSize = 1000

var IamServiceAccountSchema = map[string]*schema.Schema{
	"service_account_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The service account ID to apply a binding to.",
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

func (u *ServiceAccountIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getServiceAccountAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ServiceAccountIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	client := iamsdk.NewServiceAccountClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.serviceAccountID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMServiceAccountDefaultTimeout)
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

func (u *ServiceAccountIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	client := iamsdk.NewServiceAccountClient(u.Config.SDK)

	bSize := yandexIAMServiceAccountUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.serviceAccountID,
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

func (u *ServiceAccountIamUpdater) GetResourceID() string {
	return u.serviceAccountID
}

func (u *ServiceAccountIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-service-account-%s", u.serviceAccountID)
}

func (u *ServiceAccountIamUpdater) DescribeResource() string {
	return fmt.Sprintf("service account '%s'", u.serviceAccountID)
}

func getServiceAccountAccessBindings(ctx context.Context, config *Config, serviceAccountID string) ([]*access.AccessBinding, error) {
	client := iamsdk.NewServiceAccountClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: serviceAccountID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of service account %s: %w", serviceAccountID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
