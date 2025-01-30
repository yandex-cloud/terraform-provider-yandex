package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexDataprocCluster() *schema.Resource {
	dataSource := convertResourceToDataSource(resourceYandexDataprocCluster())
	dataSource.Schema["name"].Optional = true
	dataSource.Schema["cluster_config"].Elem.(*schema.Resource).Schema["version_id"].Optional = true

	dataSource.Schema["cluster_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Optional: true,
	}
	// TODO: SA1019: dataSource.Read is deprecated: Use ReadContext or ReadWithoutTimeout instead. This implementation does not support request cancellation initiated by Terraform, such as a system or practitioner sending SIGINT (Ctrl-c). This implementation also does not support warning diagnostics. (staticcheck)
	dataSource.Read = dataSourceYandexDataprocClusterRead
	return dataSource
}

func dataSourceYandexDataprocClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.DataprocClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Data Processing Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.Dataproc().Cluster().Get(ctx, &dataproc.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", clusterID))
	}

	d.SetId(cluster.Id)
	return populateDataprocClusterResourceData(d, config, cluster)
}
