package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexIAMServiceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexIAMServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexIAMServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	var sa *iam.ServiceAccount

	serviceAccountID := d.Get("service_account_id").(string)
	sa, err := config.sdk.IAM().ServiceAccount().Get(ctx, &iam.GetServiceAccountRequest{
		ServiceAccountId: serviceAccountID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("service account with ID %q", serviceAccountID))
	}

	createdAt, err := getTimestamp(sa.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("folder_id", sa.FolderId)
	d.Set("name", sa.Name)
	d.Set("description", sa.Description)
	d.Set("created_at", createdAt)
	d.SetId(sa.Id)

	return nil
}
