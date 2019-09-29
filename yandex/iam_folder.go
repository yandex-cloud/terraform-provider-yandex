package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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

func newFolderIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &FolderIamUpdater{
		folderID: d.Get("folder_id").(string),
		Config:   config,
	}, nil
}

func folderIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("folder_id", d.Id())
	return nil
}

func (u *FolderIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getFolderAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, err
}

func (u *FolderIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.folderID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexResourceManagerFolderDefaultTimeout)
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

func getFolderAccessBindings(config *Config, folderID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.ResourceManager().Folder().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: folderID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for folder %s: %s", folderID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
