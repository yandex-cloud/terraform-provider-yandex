package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMWorkloadIdentityOidcFederationUpdateAccessBindingsBatchSize = 1000

var IamWorkloadIdentityOidcFederationSchema = map[string]*schema.Schema{
	"federation_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "ID of the OIDC workload identity federation to apply a binding to.",
	},
}

type WorkloadIdentityOidcFederationIamUpdater struct {
	federationID string
	Config       *Config
}

func newWorkloadIdentityOidcFederationIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &WorkloadIdentityOidcFederationIamUpdater{
		federationID: d.Get("federation_id").(string),
		Config:       config,
	}, nil
}

func workloadIdentityOidcFederationIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("federation_id", d.Id())
	return nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getWorkloadIdentityOidcFederationAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.federationID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), 1*time.Minute)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.WorkloadOidc().Federation().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *WorkloadIdentityOidcFederationIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMWorkloadIdentityOidcFederationUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.federationID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.WorkloadOidc().Federation().UpdateAccessBindings(ctx, req))
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

func (u *WorkloadIdentityOidcFederationIamUpdater) GetResourceID() string {
	return u.federationID
}

func (u *WorkloadIdentityOidcFederationIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-workload-identity-oidc-federation-%s", u.federationID)
}

func (u *WorkloadIdentityOidcFederationIamUpdater) DescribeResource() string {
	return fmt.Sprintf("workload identity OIDC federation '%s'", u.federationID)
}

func getWorkloadIdentityOidcFederationAccessBindings(ctx context.Context, config *Config, federationID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.WorkloadOidc().Federation().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: federationID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of workload identity OIDC federation %s: %w", federationID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
