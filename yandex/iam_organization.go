package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexOrganizationManagerOrganizationDefaultTimeout = time.Second * 60

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

func (u *OrganizationIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getOrganizationAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *OrganizationIamUpdater) SetResourceIamPolicy(policy *Policy) error {
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
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
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

func getOrganizationAccessBindings(config *Config, organizationID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.OrganizationManager().Organization().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: organizationID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for organization %s: %s", organizationID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
