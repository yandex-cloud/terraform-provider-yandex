package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexResourceManagerCloudDefaultTimeout = time.Second * 60
const yandexResourceManagerCloudUpdateAccessBindingsBatchSize = 1000

var IamCloudSchema = map[string]*schema.Schema{
	"cloud_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type CloudIamUpdater struct {
	cloudID string
	Config  *Config
}

func newCloudIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &CloudIamUpdater{
		cloudID: d.Get("cloud_id").(string),
		Config:  config,
	}, nil
}

func cloudIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("cloud_id", d.Id())
	return nil
}

func (u *CloudIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getCloudAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *CloudIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.cloudID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexResourceManagerCloudDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ResourceManager().Cloud().SetAccessBindings(ctx, req))
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

func (u *CloudIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexResourceManagerCloudUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.cloudID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ResourceManager().Cloud().UpdateAccessBindings(ctx, req))
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

func (u *CloudIamUpdater) GetResourceID() string {
	return u.cloudID
}

func (u *CloudIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-cloud-%s", u.cloudID)
}

func (u *CloudIamUpdater) DescribeResource() string {
	return fmt.Sprintf("cloud %q", u.cloudID)
}

func getCloudAccessBindings(ctx context.Context, config *Config, cloudID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.ResourceManager().Cloud().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: cloudID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of cloud %s: %w", cloudID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
