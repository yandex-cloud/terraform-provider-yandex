package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexOrganizationManagerIAMGroupDefaultTimeout = time.Second * 60

var IamGroupSchema = map[string]*schema.Schema{
	"group_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type GroupIamUpdater struct {
	groupID string
	Config  *Config
}

func newGroupIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &GroupIamUpdater{
		groupID: d.Get("group_id").(string),
		Config:  config,
	}, nil
}

func groupIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("group_id", d.Id())
	return nil
}

func (u *GroupIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getGroupAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *GroupIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.groupID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexOrganizationManagerIAMGroupDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.OrganizationManager().Group().SetAccessBindings(ctx, req))
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

func (u *GroupIamUpdater) GetResourceID() string {
	return u.groupID
}

func (u *GroupIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-group-%s", u.groupID)
}

func (u *GroupIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Group %q", u.groupID)
}

func getGroupAccessBindings(config *Config, groupID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.OrganizationManager().Group().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: groupID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for Group %s: %s", groupID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
