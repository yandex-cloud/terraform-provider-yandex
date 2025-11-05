package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBPostgreSQLUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed PostgreSQL user. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).",

		Read: dataSourceYandexMDBPostgreSQLUserRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the PostgreSQL cluster.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the PostgreSQL user.",
				Required:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLUser().Schema["password"].Description,
				Computed:    true,
				Sensitive:   true,
			},
			"grants": {
				Type:        schema.TypeList,
				Description: resourceYandexMDBPostgreSQLUser().Schema["grants"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"login": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBPostgreSQLUser().Schema["login"].Description,
				Optional:    true,
				Default:     true,
			},
			"permission": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLUser().Schema["permission"].Description,
				Computed:    true,
				Set:         pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the database that the permission grants access to.",
						},
					},
				},
			},
			"conn_limit": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBPostgreSQLUser().Schema["conn_limit"].Description,
				Optional:    true,
			},
			"settings": {
				Type:             schema.TypeMap,
				Description:      resourceYandexMDBPostgreSQLUser().Schema["settings"].Description,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deletion_protection": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["deletion_protection"],
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"unspecified", "true", "false"}, false),
			},
			"connection_manager": {
				Type:        schema.TypeMap,
				Description: resourceYandexMDBPostgreSQLUser().Schema["connection_manager"].Description,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auth_method": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLUser().Schema["auth_method"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLUserRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)
	userID := constructResourceId(clusterID, username)
	d.SetId(userID)
	return resourceYandexMDBPostgreSQLUserRead(d, meta)
}
