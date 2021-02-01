package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMContainerRegistryDefaultTimeout = 1 * time.Minute

var IamContainerRegistrySchema = map[string]*schema.Schema{
	"registry_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type ContainerRegistryIamUpdater struct {
	registryID string
	Config     *Config
}

func newContainerRegistryIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &ContainerRegistryIamUpdater{
		registryID: d.Get("registry_id").(string),
		Config:     config,
	}, nil
}

func containerRegistryIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("registry_id", d.Id())
	return nil
}

func (u *ContainerRegistryIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getContainerRegistryAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ContainerRegistryIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.registryID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMContainerRegistryDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Registry().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	return nil
}

func (u *ContainerRegistryIamUpdater) GetResourceID() string {
	return u.registryID
}

func (u *ContainerRegistryIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-registry-%s", u.registryID)
}

func (u *ContainerRegistryIamUpdater) DescribeResource() string {
	return fmt.Sprintf("container registry '%s'", u.registryID)
}

func getContainerRegistryAccessBindings(config *Config, registryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.ContainerRegistry().Registry().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: registryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving IAM access bindings for container registry %s: %s", registryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
