package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhousesdk "github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
)

func dataSourceYandexMDBClickHouseCluster() *schema.Resource {
	dataSourceSchema := convertToOptional(resourceYandexMDBClickHouseCluster().Schema)
	removeWriteOnlyFields(dataSourceSchema)

	return &schema.Resource{
		Description:        "Get information about a Yandex Managed ClickHouse cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-clickhouse/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",
		DeprecationMessage: "The `yandex_mdb_clickhouse_cluster` data source is deprecated and will be removed in a future version. Use `yandex_mdb_clickhouse_cluster_v2` instead.",

		Read:   dataSourceYandexMDBClickHouseClusterRead,
		Schema: dataSourceSchema,
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
		folderID, err := getFolderID(d, config)
		if err != nil {
			return err
		}
		name := d.Get("name").(string)
		resolver := sdkresolversv2.NewBaseNameResolver(name, "cluster", sdkresolversv2.FolderID(folderID))
		resp, err := clickhousesdk.NewClusterClient(config.SDK).List(ctx, &clickhouse.ListClustersRequest{
			FolderId: folderID,
			Filter:   sdkresolversv2.CreateResolverFilter("name", name),
			PageSize: sdkresolversv2.DefaultResolverPageSize,
		})
		if err := resolver.FindName(resp.GetClusters(), err); err != nil {
			return fmt.Errorf("failed to resolve data source ClickHouse Cluster by name: %v", err)
		}
		clusterID = resolver.ID()

		d.Set("cluster_id", clusterID)
	}

	d.SetId(clusterID)
	return resourceYandexMDBClickHouseClusterRead(d, meta)
}
