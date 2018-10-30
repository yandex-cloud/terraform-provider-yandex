package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

var IamFolderSchema = map[string]*schema.Schema{
	"folder_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type FolderIamUpdater struct {
	folderID string
	Config   *Config
}

func NewFolderIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &FolderIamUpdater{
		folderID: d.Get("folder_id").(string),
		Config:   config,
	}, nil
}

func FolderIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("folder_id", d.Id())
	return nil
}

func (u *FolderIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	return getFolderIamPolicyByFolderID(u.folderID, u.Config)
}

func (u *FolderIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.folderID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(context.Background(), yandexResourceManagerFolderDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ResourceManager().Folder().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil
}

func (u *FolderIamUpdater) GetResourceID() string {
	return u.folderID
}

func (u *FolderIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-folder-%s", u.folderID)
}

func (u *FolderIamUpdater) DescribeResource() string {
	return fmt.Sprintf("folder %q", u.folderID)
}

// Retrieve the existing IAM Policy for a folder
func getFolderIamPolicyByFolderID(folderID string, config *Config) (*Policy, error) {
	resp, err := config.sdk.ResourceManager().Folder().ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
		ResourceId: folderID,
	})

	if err != nil {
		return nil, fmt.Errorf("Error retrieving IAM policy for folder %q: %s", folderID, err)
	}

	return &Policy{resp.AccessBindings}, nil
}
