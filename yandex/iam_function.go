package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMFunctionDefaultTimeout = 1 * time.Minute

// const yandexFunctionUpdateAccessBindingsBatchSize = 1000 // TODO: deadcode

var IamFunctionSchema = map[string]*schema.Schema{
	"function_id": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
}

type FunctionIamUpdater struct {
	functionID string
	Config     *Config
}

func newFunctionIamUpdater(d *schema.ResourceData, config *Config) (ResourceIamUpdater, error) {
	return &FunctionIamUpdater{
		functionID: d.Get("function_id").(string),
		Config:     config,
	}, nil
}

func functionIDParseFunc(d *schema.ResourceData, _ *Config) error {
	d.Set("function_id", d.Id())
	return nil
}

func (u *FunctionIamUpdater) GetResourceIamPolicy(ctx context.Context) (*Policy, error) {
	bindings, err := getFunctionAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *FunctionIamUpdater) SetResourceIamPolicy(ctx context.Context, policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.functionID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, yandexIAMFunctionDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Serverless().Functions().Function().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *FunctionIamUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *PolicyDelta) error {
	bSize := yandexResourceManagerCloudUpdateAccessBindingsBatchSize
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < countBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          u.functionID,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}

		op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Serverless().Functions().Function().UpdateAccessBindings(ctx, req))
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

func (u *FunctionIamUpdater) GetResourceID() string {
	return u.functionID
}

func (u *FunctionIamUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-function-%s", u.functionID)
}

func (u *FunctionIamUpdater) DescribeResource() string {
	return fmt.Sprintf("function '%s'", u.functionID)
}

func getFunctionAccessBindings(config *Config, functionID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""
	ctx := config.Context()

	for {
		resp, err := config.sdk.Serverless().Functions().Function().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: functionID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of function %s: %w", functionID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
