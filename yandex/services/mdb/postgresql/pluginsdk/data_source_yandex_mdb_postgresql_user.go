package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func DataSourceYandexMDBPostgreSQLUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed PostgreSQL user. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).",

		Read: dataSourceYandexMDBPostgreSQLUserRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["cluster_id"].Description,
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["name"].Description,
				Required:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["password"].Description,
				Computed:    true,
				Sensitive:   true,
			},
			"grants": {
				Type:        schema.TypeList,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["grants"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"login": {
				Type:        schema.TypeBool,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["login"].Description,
				Computed:    true,
			},
			"permission": {
				Type:        schema.TypeSet,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["permission"].Description,
				Computed:    true,
				Set:         pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: ResourceYandexMDBPostgreSQLUser().Schema["permission"].Elem.(*schema.Resource).Schema["database_name"].Description,
						},
					},
				},
			},
			"conn_limit": {
				Type:        schema.TypeInt,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["conn_limit"].Description,
				Computed:    true,
			},
			"settings": {
				Type:        schema.TypeMap,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["settings"].Description,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deletion_protection": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
			},
			"connection_manager": {
				Type:        schema.TypeMap,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["connection_manager"].Description,
				Deprecated:  sdkcommon.FieldDeprecatedForAnother("connection_manager", "user_connection_manager"),
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_connection_manager": mdbcommon.UserConnectionManagerDataSourceSchema(),
			"auth_method": {
				Type:        schema.TypeString,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["auth_method"].Description,
				Computed:    true,
			},
			"user_password_encryption": {
				Type:        schema.TypeString,
				Description: ResourceYandexMDBPostgreSQLUser().Schema["user_password_encryption"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLUserRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)
	userID := sdkcommon.ConstructResourceId(clusterID, username)
	d.SetId(userID)
	return resourceYandexMDBPostgreSQLUserRead(d, meta)
}
