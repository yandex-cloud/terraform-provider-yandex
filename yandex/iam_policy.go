package yandex

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func iamPolicyImport(resourceIDParser resourceIDParserFunc) schema.StateFunc {
	return func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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

func resourceIamPolicy(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc) *schema.Resource {
	return &schema.Resource{
		Create: resourceIamPolicyCreate(newUpdaterFunc),
		Read:   resourceIamPolicyRead(newUpdaterFunc),
		Update: resourceIamPolicyUpdate(newUpdaterFunc),
		Delete: resourceIamPolicyDelete(newUpdaterFunc),

		Schema: mergeSchemas(IamPolicyBaseSchema, parentSpecificSchema),
	}
}

func resourceIamPolicyWithImport(parentSpecificSchema map[string]*schema.Schema, newUpdaterFunc newResourceIamUpdaterFunc, resourceIDParser resourceIDParserFunc) *schema.Resource {
	r := resourceIamPolicy(parentSpecificSchema, newUpdaterFunc)
	r.Importer = &schema.ResourceImporter{
		State: iamPolicyImport(resourceIDParser),
	}
	return r
}

func resourceIamPolicyCreate(newUpdaterFunc newResourceIamUpdaterFunc) schema.CreateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		if err := setIamPolicyData(d, updater); err != nil {
			return err
		}

		d.SetId(updater.GetResourceID())
		return resourceIamPolicyRead(newUpdaterFunc)(d, meta)
	}
}

func resourceIamPolicyRead(newUpdaterFunc newResourceIamUpdaterFunc) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		policy, err := updater.GetResourceIamPolicy()

		if err != nil {
			if isStatusWithCode(err, codes.NotFound) {
				log.Printf("[DEBUG]: Policy does not exist for non-existent resource %q", updater.GetResourceID())
				return nil
			}
			return err
		}

		d.Set("policy_data", marshalIamPolicy(policy))

		return nil
	}
}

func resourceIamPolicyUpdate(newUpdaterFunc newResourceIamUpdaterFunc) schema.UpdateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		if d.HasChange("policy_data") {
			if err := setIamPolicyData(d, updater); err != nil {
				return err
			}
		}

		return resourceIamPolicyRead(newUpdaterFunc)(d, meta)
	}
}

func resourceIamPolicyDelete(newUpdaterFunc newResourceIamUpdaterFunc) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		config := meta.(*Config)
		updater, err := newUpdaterFunc(d, config)
		if err != nil {
			return err
		}

		// Set an empty policy to delete the attached policy.
		err = updater.SetResourceIamPolicy(&Policy{})
		return err
	}
}

func setIamPolicyData(d *schema.ResourceData, updater ResourceIamUpdater) error {
	policy, err := unmarshalIamPolicy(d.Get("policy_data").(string))
	if err != nil {
		return fmt.Errorf("'policy_data' is not valid for %s: %s", updater.DescribeResource(), err)
	}

	err = updater.SetResourceIamPolicy(policy)
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
		return nil, fmt.Errorf("Could not unmarshal policy data %s:\n%s", policyData, err)
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
