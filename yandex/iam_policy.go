package yandex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"
)

var IamPolicyBaseSchema = map[string]*schema.Schema{
	"policy_data": {
		Type:             schema.TypeString,
		Required:         true,
		DiffSuppressFunc: shouldSuppressDiffForPolicies,
		ValidateFunc:     validateIamPolicy,
	},
}

func iamPolicyImport(resourceIDParser resourceIDParserFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
		if resourceIDParser == nil {
			return nil, errors.New("Import not supported for this IAM resource")
		}
		config := m.(*Config)
		err := resourceIDParser(d, config)
		if err != nil {
			return nil, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

func resourceIamPolicy(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, opts ...SchemaOption) *schema.Resource {
	r := &schema.Resource{
		CreateContext: resourceIamPolicyCreate(newUpdaterFunc),
		ReadContext:   resourceIamPolicyRead(newUpdaterFunc),
		UpdateContext: resourceIamPolicyUpdate(newUpdaterFunc),
		DeleteContext: resourceIamPolicyDelete(newUpdaterFunc),

		Schema: mergeSchemas(IamPolicyBaseSchema, parentSpecificSchema),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func resourceIamPolicyCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := setIamPolicyData(ctx, d, updater); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(updater.GetResourceID())
		return resourceIamPolicyRead(newUpdaterFunc)(ctx, d, meta)
	}
}

func resourceIamPolicyRead(newUpdaterFunc newResourceIamUpdaterFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		policy, err := updater.GetResourceIamPolicy(ctx)

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Acccess bindings not exist for non-existent resource %q", updater.GetResourceID())
				return nil
			}
			return diag.FromErr(err)
		}

		d.Set("policy_data", marshalIamPolicy(policy))

		return nil
	}
}

func resourceIamPolicyUpdate(newUpdaterFunc newResourceIamUpdaterFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("policy_data") {
			if err := setIamPolicyData(ctx, d, updater); err != nil {
				return diag.FromErr(err)
			}
		}

		return resourceIamPolicyRead(newUpdaterFunc)(ctx, d, meta)
	}
}

func resourceIamPolicyDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return diag.FromErr(err)
		}

		// Set an empty policy to delete the attached policy.
		err = updater.SetResourceIamPolicy(ctx, &Policy{})
		return diag.FromErr(err)
	}
}

func setIamPolicyData(ctx context.Context, d *schema.ResourceData, updater ResourceIamUpdater) error {
	policy, err := unmarshalIamPolicy(d.Get("policy_data").(string))
	if err != nil {
		return fmt.Errorf("'policy_data' is not valid for %s: %w", updater.DescribeResource(), err)
	}

	err = updater.SetResourceIamPolicy(ctx, policy)
	return err
}

func marshalIamPolicy(policy *Policy) string {
	pdBytes, _ := json.Marshal(&Policy{
		Bindings: policy.Bindings,
	})

	return string(pdBytes)
}

func unmarshalIamPolicy(policyData string) (*Policy, error) {
	policy := &Policy{}
	if err := json.Unmarshal([]byte(policyData), policy); err != nil {
		return nil, fmt.Errorf("Could not unmarshal policy data %s:\n%w", policyData, err)
	}
	return policy, nil
}

func validateIamPolicy(i interface{}, k string) (s []string, es []error) {
	_, err := unmarshalIamPolicy(i.(string))
	if err != nil {
		es = append(es, err)
	}
	return
}
