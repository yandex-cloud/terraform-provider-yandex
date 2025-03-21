package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexMDBMySQLUser() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed MySQL user. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mysql/).",

		Read: dataSourceYandexMDBMySQLUserRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the MySQL cluster.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the MySQL user.",
				Required:    true,
			},
			"permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      mysqlUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"roles": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
						},
					},
				},
			},
			"global_permissions": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBMySQLUser().Schema["global_permissions"].Description,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"connection_limits": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_questions_per_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_updates_per_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_connections_per_hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_user_connections": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"authentication_plugin": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBMySQLUser().Schema["authentication_plugin"].Description,
				Computed:    true,
			},
			"connection_manager": {
				Type:        schema.TypeMap,
				Description: resourceYandexMDBMySQLUser().Schema["connection_manager"].Description,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceYandexMDBMySQLUserRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)
	userID := constructResourceId(clusterID, username)
	d.SetId(userID)
	return resourceYandexMDBMySQLUserRead(d, meta)
}
