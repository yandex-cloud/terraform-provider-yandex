package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const YandexOrganizationManagerIAMGroupDefaultTimeout = time.Second * 60
const yandexOrganizationManagerIAMGroupUpdateAccessBindingsBatchSize = 1000

var IamGroupSchema = map[string]*schema.Schema{
	"group_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The ID of the organization to attach a policy to.",
	},
}

type GroupIamUpdater struct {
	groupID string
	Config  *providerconfig.Config
}

func newGroupIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return &GroupIamUpdater{
		groupID: d.Get("group_id").(string),
		Config:  config,
	}, nil
}

func groupIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("group_id", d.Id())
	return nil
}

func (u *GroupIamUpdater) GetResourceIamPolicy(ctx context.Context) (*sdkcommon.Policy, error) {
	bindings, err := GetGroupAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &sdkcommon.Policy{bindings}, nil
}

func (u *GroupIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *sdkcommon.Policy) error {
	client := organizationmanagersdk.NewGroupClient(u.Config.SDK)

	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.groupID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, YandexOrganizationManagerIAMGroupDefaultTimeout)
	defer cancel()

	op, err := client.SetAccessBindings(ctx, req)
	if err != nil {
		if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *GroupIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *sdkcommon.PolicyDelta) error {
	client := organizationmanagersdk.NewGroupClient(u.Config.SDK)

	bSize := yandexOrganizationManagerIAMGroupUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < sdkcommon.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.groupID,
			AccessBindingDeltas: deltas[i*bSize : sdkcommon.Min((i+1)*bSize, dLen)],
		}

		op, err := client.UpdateAccessBindings(ctx, req)
		if err != nil {
			if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
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

func (u *GroupIamUpdater) GetResourceID() string {
	return u.groupID
}

func (u *GroupIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-group-%s", u.groupID)
}

func (u *GroupIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Group %q", u.groupID)
}

func GetGroupAccessBindings(ctx context.Context, config *providerconfig.Config, groupID string) ([]*access.AccessBinding, error) {
	client := organizationmanagersdk.NewGroupClient(config.SDK)

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := client.ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: groupID,
			PageSize:   sdkcommon.DefaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for Group %s: %w", groupID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
