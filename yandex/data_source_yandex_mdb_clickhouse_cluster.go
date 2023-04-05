package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceYandexMDBClickHouseClusterRead,
		Schema: convertToOptional(resourceYandexMDBClickHouseCluster().Schema),
	}
}

func dataSourceYandexMDBClickHouseClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.ClickhouseClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source ClickHouse Cluster by name: %v", err)
		}

		d.Set("cluster_id", clusterID)
	}

	d.SetId(clusterID)
	return resourceYandexMDBClickHouseClusterRead(d, meta)
}
