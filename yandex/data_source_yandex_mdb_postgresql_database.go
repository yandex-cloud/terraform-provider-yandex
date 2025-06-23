package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBPostgreSQLDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed PostgreSQL database. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).",

		Read: dataSourceYandexMDBPostgreSQLDatabaseRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["cluster_id"].Description,
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"owner": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["owner"].Description,
				Computed:    true,
				Optional:    true,
			},
			"lc_collate": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["lc_collate"].Description,
				Computed:    true,
				Optional:    true,
			},
			"lc_type": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["lc_type"].Description,
				Computed:    true,
				Optional:    true,
			},
			"template_db": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["template_db"].Description,
				Computed:    true,
				Optional:    true,
			},
			"extension": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBPostgreSQLDatabase().Schema["extension"].Description,
				Set:         pgExtensionHash,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "Name of the database extension. For more information on available extensions see [the official documentation](https://yandex.cloud/docs/managed-postgresql/operations/cluster-extensions).",
							Required:    true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["deletion_protection"],
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"unspecified", "true", "false"}, false),
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
