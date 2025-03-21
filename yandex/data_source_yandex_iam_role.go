package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexIAMRole() *schema.Resource {
	return &schema.Resource{
		Description: "Generates an [IAM](https://yandex.cloud/docs/iam/) role document that may be referenced by and applied to other Yandex Cloud Platform resources, such as the `yandex_resourcemanager_folder` resource. For more information, see [the official documentation](https://yandex.cloud/docs/iam/concepts/access-control/roles).\n\nThis data source is used to define [IAM](https://yandex.cloud/docs/iam/) roles in order to apply them to other resources. Currently, defining a role through a data source and referencing that role from another resource is the only way to apply an IAM role to a resource.\n\n",
		Read:        dataSourceYandexIAMRoleRead,
		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceYandexIAMRoleRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	var role *iam.Role

	v, ok := d.GetOk("role_id")
	if !ok {
		return fmt.Errorf("'role_id' must be set")
	}

	resp, err := config.sdk.IAM().Role().Get(ctx, &iam.GetRoleRequest{
		RoleId: v.(string),
	})

	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			return fmt.Errorf("role not found: %s", v)
		}
		return err
	}

	role = resp

	d.SetId(role.Id)
	d.Set("description", role.Description)

	return nil
}
