package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexMDBPostgreSQLUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBPostgreSQLUserRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"grants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"permission": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"conn_limit": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"settings": {
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
