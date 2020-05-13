package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const yandexIAMFunctionDefaultTimeout = 1 * time.Minute

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

func (u *FunctionIamUpdater) GetResourceIamPolicy() (*Policy, error) {
	bindings, err := getFunctionAccessBindings(u.Config, u.GetResourceID())
	if err != nil {
		return nil, err
	}
	return &Policy{bindings}, nil
}

func (u *FunctionIamUpdater) SetResourceIamPolicy(policy *Policy) error {
	req := &access.SetAccessBindingsRequest{
		ResourceId:     u.functionID,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(u.Config.Context(), yandexIAMFunctionDefaultTimeout)
	defer cancel()

	op, err := u.Config.sdk.WrapOperation(u.Config.sdk.Serverless().Functions().Function().SetAccessBindings(ctx, req))
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error setting IAM policy for %s: %s", u.DescribeResource(), err)
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
			return nil, fmt.Errorf("Error retrieving IAM access bindings for function %s: %s", functionID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
