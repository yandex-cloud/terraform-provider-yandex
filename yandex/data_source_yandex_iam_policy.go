package yandex

import (
	"encoding/json"
	"strconv"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceYandexIamPolicy returns a *schema.Resource that allows a customer
// to express a Yandex Cloud IAM policy in a data resource. This is an example
// of how the schema would be used in a config:
//
// data "yandex_iam_policy" "admin" {
//   binding {
//     role = "roles/viewer"
//     members = [
//       "userAccount:some_user_id",
//     ]
//   }
// }

var iamBinding = &schema.Schema{
	Type:     schema.TypeSet,
	Required: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"members": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIamMember,
				},
				Set: schema.HashString,
			},
		},
	},
}

func dataSourceYandexIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"binding": iamBinding,
			"policy_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Read: dataSourceYandexIAMPolicyRead,
	}
}

func dataSourceYandexIAMPolicyRead(d *schema.ResourceData, meta interface{}) error {
	var policy Policy

	// The schema supports multiple binding{} blocks
	bset := d.Get("binding").(*schema.Set)

	// Convert each config binding into a access.AccessBinding
	for _, v := range bset.List() {
		binding := v.(map[string]interface{})
		role := binding["role"].(string)
		for _, member := range convertStringSet(binding["members"].(*schema.Set)) {
			policy.Bindings = append(policy.Bindings, roleMemberToAccessBinding(role, member))
		}
	}

	// Marshal Policy to JSON suitable for storing in state
	jsonPolicy, err := json.Marshal(&policy)
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	stringPolicy := string(jsonPolicy)

	d.Set("policy_data", stringPolicy)
	d.SetId(strconv.Itoa(hashcode.String(stringPolicy))) // TODO: SA1019: hashcode.String is deprecated: This will be removed in v2 without replacement. If you need its functionality, you can copy it, import crc32 directly, or reference the v1 package. (staticcheck)

	return nil

}
