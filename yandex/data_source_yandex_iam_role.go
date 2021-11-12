package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexIAMRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexIAMRoleRead,
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
