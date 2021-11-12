package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMContainerRepositoryDefaultTimeout = 1 * time.Minute

var IamContainerRepositorySchema = map[string]*schema.Schema{
	"repository_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type ContainerRepositoryIamUpdater struct {
	repositoryID string
	Config       *Config
}

func newContainerRepositoryIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &ContainerRepositoryIamUpdater{
		repositoryID: d.Get("repository_id").(string),
		Config:       config,
	}, nil
}

func containerRepositoryIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("repository_id", d.Id())
	return nil
}

func (u *ContainerRepositoryIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getContainerRepositoryAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ContainerRepositoryIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.repositoryID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMContainerRepositoryDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Repository().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil
}

func (u *ContainerRepositoryIamUpdater) GetResourceID() string {
	return u.repositoryID
}

func (u *ContainerRepositoryIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-repository-%s", u.repositoryID)
}

func (u *ContainerRepositoryIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Container Repository '%s'", u.repositoryID)
}

func getContainerRepositoryAccessBindings(config *Config, repositoryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.ContainerRegistry().Repository().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: repositoryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for Container Repository %s: %s", repositoryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
