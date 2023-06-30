package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMContainerRepositoryDefaultTimeout = 1 * time.Minute
const yandexIAMContainerRepositoryUpdateAccessBindingsBatchSize = 1000

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

func (u *ContainerRepositoryIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getContainerRepositoryAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ContainerRepositoryIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.repositoryID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMContainerRepositoryDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Repository().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *ContainerRepositoryIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMContainerRepositoryUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.repositoryID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.ContainerRegistry().Repository().UpdateAccessBindings(ctx, req))
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

func (u *ContainerRepositoryIamUpdater) GetResourceID() string {
	return u.repositoryID
}

func (u *ContainerRepositoryIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-repository-%s", u.repositoryID)
}

func (u *ContainerRepositoryIamUpdater) DescribeResource() string {
	return fmt.Sprintf("Container Repository '%s'", u.repositoryID)
}

func getContainerRepositoryAccessBindings(ctx context.Context, config *Config, repositoryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.ContainerRegistry().Repository().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: repositoryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of Container Repository %s: %w", repositoryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
