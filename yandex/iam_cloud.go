package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexResourceManagerCloudDefaultTimeout = time.Second * 60

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

func (u *CloudIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getCloudAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *CloudIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.cloudID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexResourceManagerCloudDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ResourceManager().Cloud().SetAccessBindings(ctx, req))
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

func (u *CloudIamUpdater) GetResourceID() string {
	return u.cloudID
}

func (u *CloudIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-cloud-%s", u.cloudID)
}

func (u *CloudIamUpdater) DescribeResource() string {
	return fmt.Sprintf("cloud %q", u.cloudID)
}

func getCloudAccessBindings(config *Config, cloudID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.ResourceManager().Cloud().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: cloudID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for cloud %s: %s", cloudID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
