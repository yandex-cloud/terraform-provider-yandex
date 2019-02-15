package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"

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
	resourceID string
	Config     *Config
}

func newCloudIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &CloudIamUpdater{
		resourceID: d.Get("cloud_id").(string),
		Config:     config,
	}, nil
}

func cloudIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("cloud_id", d.Id())
	return nil
}

func (u *CloudIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	return getCloudIamPolicyByCloudID(u.resourceID, u.Config)
}

func (u *CloudIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.resourceID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(context.Background(), yandexResourceManagerCloudDefaultTimeout)
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
	return u.resourceID
}

func (u *CloudIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-cloud-%s", u.resourceID)
}

func (u *CloudIamUpdater) DescribeResource() string {
	return fmt.Sprintf("cloud %q", u.resourceID)
}

// Retrieve the existing IAM Policy for a cloud
func getCloudIamPolicyByCloudID(cloudID string, config *Config) (*Policy, error) {
	resp, err := config.sdk.ResourceManager().Cloud().ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
		ResourceId: cloudID,
	})

	if err != nil {
		return nil, fmt.Errorf("Error retrieving IAM policy for cloud %q: %s", cloudID, err)
	}

	return &Policy{resp.AccessBindings}, nil
}
