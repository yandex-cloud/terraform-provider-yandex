package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexMDBPostgreSQLDatabase() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBPostgreSQLDatabaseRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"lc_collate": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "C",
			},
			"lc_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "C",
			},
			"extension": {
				Type:     schema.TypeSet,
				Set:      pgExtensionHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	dbname := d.Get("name").(string)
	databaseID := constructResourceId(clusterID, dbname)
	d.SetId(databaseID)
	return resourceYandexMDBPostgreSQLDatabaseRead(d, meta)
}
