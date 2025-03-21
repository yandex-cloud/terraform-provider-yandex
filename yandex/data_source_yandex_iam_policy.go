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
	Type:        schema.TypeSet,
	Description: "Defines a binding to be included in the policy document. Multiple `binding` arguments are supported.",
	Required:    true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role": {
				Type:        schema.TypeString,
				Description: "The role/permission that will be granted to the members. See the [IAM Roles](https://yandex.cloud/docs/iam/concepts/access-control/roles) documentation for a complete list of roles.",
				Required:    true,
			},
			"members": {
				Type:        schema.TypeSet,
				Description: "An array of identities that will be granted the privilege in the `role`. Each entry can have one of the following values:\n* **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.\n* **serviceAccount:{service_account_id}**: A unique service account ID.\n* **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.\n* **group:{group_id}**: A unique group ID.\n* **system:group:federation:{federation_id}:users**: All users in federation.\n* **system:group:organization:{organization_id}:users**: All users in organization.\n* **system:allAuthenticatedUsers**: All authenticated users.\n* **system:allUsers**: All users, including unauthenticated ones.\n\n~> For more information about system groups, see the [documentation](https://yandex.cloud/docs/iam/concepts/access-control/system-group).\n",
				Required:    true,
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
		Description: "Generates an [IAM](https://yandex.cloud/docs/iam/) policy document that may be referenced by and applied to other Yandex Cloud Platform resources, such as the `yandex_resourcemanager_folder` resource.\n\nThis data source is used to define [IAM](https://yandex.cloud/docs/iam/) policies to apply to other resources. Currently, defining a policy through a data source and referencing that policy from another resource is the only way to apply an IAM policy to a resource.\n",
		Schema: map[string]*schema.Schema{
			"binding": iamBinding,
			"policy_data": {
				Type:        schema.TypeString,
				Description: "The above bindings serialized in a format suitable for referencing from a resource that supports IAM.",
				Computed:    true,
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
