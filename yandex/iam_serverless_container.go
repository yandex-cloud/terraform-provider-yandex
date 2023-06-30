package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMServerlessContainerDefaultTimeout = 1 * time.Minute
const yandexIAMServerlessContainerUpdateAccessBindingsBatchSize = 1000

var IamServerlessContainerSchema = map[string]*schema.Schema{
	"container_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type ServerlessContainerIamUpdater struct {
	containerID string
	Config      *Config
}

func newServerlessContainerIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &ServerlessContainerIamUpdater{
		containerID: d.Get("container_id").(string),
		Config:      config,
	}, nil
}

func serverlessContainerIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("container_id", d.Id())
	return nil
}

func (u *ServerlessContainerIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getServerlessContainerAccessBindings(ctx, u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *ServerlessContainerIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.containerID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMServerlessContainerDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Serverless().Containers().Container().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *ServerlessContainerIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexIAMServerlessContainerUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.containerID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Serverless().Containers().Container().UpdateAccessBindings(ctx, req))
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

func (u *ServerlessContainerIamUpdater) GetResourceID() string {
	return u.containerID
}

func (u *ServerlessContainerIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-container-%s", u.containerID)
}

func (u *ServerlessContainerIamUpdater) DescribeResource() string {
	return fmt.Sprintf("container '%s'", u.containerID)
}

func getServerlessContainerAccessBindings(ctx context.Context, config *Config, containerID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.sdk.Serverless().Containers().Container().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: containerID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of container %s: %w", containerID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
