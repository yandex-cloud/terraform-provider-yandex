package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexIAMServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex IAM service account. For more information about accounts, see [Yandex Cloud IAM accounts](https://yandex.cloud/docs/iam/concepts/#accounts).\n\n~> One of `service_account_id` or `name` should be specified.\n",

		Read: dataSourceYandexIAMServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:          schema.TypeString,
				Description:   "ID of a specific service account.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:          schema.TypeString,
				Description:   common.ResourceDescriptions["name"],
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"service_account_id"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexIAMServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	var sa *iam.ServiceAccount

	err := checkOneOf(d, "service_account_id", "name")
	if err != nil {
		return err
	}

	serviceAccountID := d.Get("service_account_id").(string)
	_, serviceAccountNameOk := d.GetOk("name")

	if serviceAccountNameOk {
		serviceAccountID, err = resolveObjectID(ctx, config, d, sdkresolvers.ServiceAccountResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve service account by name: %v", err)
		}
	}

	sa, err = config.sdk.IAM().ServiceAccount().Get(ctx, &iam.GetServiceAccountRequest{
		ServiceAccountId: serviceAccountID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("service account with ID %q", serviceAccountID))
	}

	d.Set("service_account_id", sa.Id)
	d.Set("folder_id", sa.FolderId)
	d.Set("name", sa.Name)
	d.Set("description", sa.Description)
	d.Set("created_at", getTimestamp(sa.CreatedAt))
	d.SetId(sa.Id)

	return nil
}
