package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexMDBMySQLUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBMySQLUserRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
				Type: schema.TypeSet,
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
				Type:     schema.TypeString,
				Computed: true,
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
