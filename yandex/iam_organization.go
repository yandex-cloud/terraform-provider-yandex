package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexOrganizationManagerOrganizationDefaultTimeout = time.Second * 60
const yandexOrganizationManagerOrganizationUpdateAccessBindingsBatchSize = 1000

var IamOrganizationSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type OrganizationIamUpdater struct {
	organizationID string
	Config         *Config
}

func newOrganizationIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &OrganizationIamUpdater{
		organizationID: d.Get("organization_id").(string),
		Config:         config,
	}, nil
}

func organizationIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("organization_id", d.Id())
	return nil
}

func (u *OrganizationIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getOrganizationAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *OrganizationIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.organizationID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexOrganizationManagerOrganizationDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.OrganizationManager().Organization().SetAccessBindings(ctx, req))
	if err != nil {
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *OrganizationIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexOrganizationManagerOrganizationUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.organizationID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.OrganizationManager().Organization().UpdateAccessBindings(ctx, req))
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

func (u *OrganizationIamUpdater) GetResourceID() string {
	return u.organizationID
}

func (u *OrganizationIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-organization-%s", u.organizationID)
}

func (u *OrganizationIamUpdater) DescribeResource() string {
	return fmt.Sprintf("organization %q", u.organizationID)
}

func getOrganizationAccessBindings(ctx context.Context, config *Config, organizationID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.OrganizationManager().Organization().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: organizationID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of  organization %s: %w", organizationID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
