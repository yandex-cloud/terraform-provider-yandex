package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMContainerRegistryDefaultTimeout = 1 * time.Minute
const yandexContainerRegistryUpdateAccessBindingsBatchSize = 1000

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

func (u *ContainerRegistryIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getContainerRegistryAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ContainerRegistryIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.registryID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMContainerRegistryDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Registry().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *ContainerRegistryIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexContainerRegistryUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.registryID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Registry().UpdateAccessBindings(ctx, req))
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

func (u *ContainerRegistryIamUpdater) GetResourceID() string {
	return u.registryID
}

func (u *ContainerRegistryIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-registry-%s", u.registryID)
}

func (u *ContainerRegistryIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Container Registry '%s'", u.registryID)
}

func getContainerRegistryAccessBindings(ctx context.Context, config *Config, registryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.ContainerRegistry().Registry().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: registryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of Container Registry %s: %w", registryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
