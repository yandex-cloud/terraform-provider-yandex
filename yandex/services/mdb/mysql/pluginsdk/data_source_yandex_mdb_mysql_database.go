package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func DataSourceYandexMDBMySQLDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Get details from MySQL database at MySQL Cluster.",

		Read: dataSourceYandexMDBMySQLDatabaseRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The MySQL cluster ID.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the database.",
				Required:    true,
			},
		},
	}
}

func dataSourceYandexMDBMySQLDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	clusterID := d.Get("cluster_id").(string)
	dbname := d.Get("name").(string)
	databaseID := sdkcommon.ConstructResourceId(clusterID, dbname)
	d.SetId(databaseID)
	return resourceYandexMDBMySQLDatabaseRead(d, meta)
}
