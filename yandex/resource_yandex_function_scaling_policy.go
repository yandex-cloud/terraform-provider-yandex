package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

func resourceYandexFunctionScalingPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexFunctionScalingPolicyCreate,
		Read:   resourceYandexFunctionScalingPolicyRead,
		Update: resourceYandexFunctionScalingPolicyUpdate,
		Delete: resourceYandexFunctionScalingPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceYandexFunctionScalingPolicyImporterFunc,
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"policy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"zone_requests_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"zone_instances_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
					},
				},
			},
		},
	}
}

func resourceYandexFunctionScalingPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	err := compareAndUpdateFunctionScalingPolicies(nil, d.Get("policy").(*schema.Set), d, meta)
	if err != nil {
		return err
	}

	functionID := d.Get("function_id").(string)
	d.SetId(functionID)

	return resourceYandexFunctionScalingPolicyRead(d, meta)
}

func resourceYandexFunctionScalingPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	functionID := d.Get("function_id").(string)

	policies, err := fetchFunctionScalingPolicies(ctx, config, functionID)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %s Scaling Policy", functionID))
	}

	return flattenYandexFunctionScalingPolicy(d, policies)
}

func resourceYandexFunctionScalingPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("policy") {
		o, n := d.GetChange("policy")
		err := compareAndUpdateFunctionScalingPolicies(o.(*schema.Set), n.(*schema.Set), d, meta)
		if err != nil {
			return err
		}
	}

	return resourceYandexFunctionScalingPolicyRead(d, meta)
}

func resourceYandexFunctionScalingPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	return compareAndUpdateFunctionScalingPolicies(d.Get("policy").(*schema.Set), nil, d, meta)
}

func resourceYandexFunctionScalingPolicyImporterFunc(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("function_id", d.Id())

	return []*schema.ResourceData{d}, nil
}

func fetchFunctionScalingPolicies(ctx context.Context, config *Config, functionID string) ([]*functions.ScalingPolicy, error) {
	var policies []*functions.ScalingPolicy
	var nextPageToken = ""

	for {
		req := &functions.ListScalingPoliciesRequest{
			FunctionId: functionID,
			PageToken:  nextPageToken,
		}

		resp, err := config.sdk.Serverless().Functions().Function().ListScalingPolicies(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("Cannot fetch Yandex Cloud Function Scaling Policies: %s", err)
		}

		policies = append(policies, resp.ScalingPolicies...)

		nextPageToken = resp.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return policies, nil
}

func expandFunctionScalingPolicies(set *schema.Set) (*map[string]*functions.ScalingPolicy, error) {
	policies := make(map[string]*functions.ScalingPolicy)

	if set != nil {
		for _, element := range set.List() {
			policy := element.(map[string]interface{})
			tag := policy["tag"].(string)
			_, ok := policies[tag]
			if ok {
				return nil, fmt.Errorf("Duplicated Yandex Cloud Function Scaling Policy with tag %s", tag)
			}
			policies[tag] = &functions.ScalingPolicy{
				ZoneRequestsLimit:  int64(policy["zone_requests_limit"].(int)),
				ZoneInstancesLimit: int64(policy["zone_instances_limit"].(int)),
			}
		}
	}

	return &policies, nil
}

func flattenYandexFunctionScalingPolicy(d *schema.ResourceData, scalingPolicies []*functions.ScalingPolicy) error {
	var policies []map[string]interface{}
	for _, policy := range scalingPolicies {
		m := make(map[string]interface{})
		m["tag"] = policy.Tag
		m["zone_requests_limit"] = policy.ZoneRequestsLimit
		m["zone_instances_limit"] = policy.ZoneInstancesLimit
		policies = append(policies, m)
	}
	return d.Set("policy", policies)
}

func compareAndUpdateFunctionScalingPolicies(oldPoliciesSet *schema.Set, newPoliciesSet *schema.Set, d *schema.ResourceData, meta interface{}) error {
	newPolicies, err := expandFunctionScalingPolicies(newPoliciesSet)
	if err != nil {
		return err
	}
	oldPolicies, err := expandFunctionScalingPolicies(oldPoliciesSet)
	if err != nil {
		return err
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	functionID := d.Get("function_id").(string)

	for tag, newPolicy := range *newPolicies {
		oldPolicy := (*oldPolicies)[tag]
		if oldPolicy != newPolicy {
			req := &functions.SetScalingPolicyRequest{
				FunctionId:         functionID,
				Tag:                tag,
				ZoneInstancesLimit: newPolicy.ZoneInstancesLimit,
				ZoneRequestsLimit:  newPolicy.ZoneRequestsLimit,
			}

			op, err := config.sdk.Serverless().Functions().Function().SetScalingPolicy(ctx, req)
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Error while requesting API to set Yandex Cloud Function Scaling Policy: %s", err)
			}
		}
	}

	for tag := range *oldPolicies {
		_, ok := (*newPolicies)[tag]
		if !ok {
			req := &functions.RemoveScalingPolicyRequest{
				FunctionId: functionID,
				Tag:        tag,
			}

			op, err := config.sdk.Serverless().Functions().Function().RemoveScalingPolicy(ctx, req)
			err = waitOperation(ctx, config, op, err)
			if err != nil {
				return fmt.Errorf("Error while requesting API to remove Yandex Cloud Function Scaling Policy: %s", err)
			}
		}
	}

	return nil
}
