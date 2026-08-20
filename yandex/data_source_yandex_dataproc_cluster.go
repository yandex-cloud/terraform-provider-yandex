package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	dataprocsdk "github.com/yandex-cloud/go-sdk/services/dataproc/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
)

func dataSourceYandexDataprocCluster() *schema.Resource {
	dataSource := convertResourceToDataSource(resourceYandexDataprocCluster())
	dataSource.Schema["name"].Optional = true
	dataSource.Schema["cluster_config"].Elem.(*schema.Resource).Schema["version_id"].Optional = true

	dataSource.Schema["cluster_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The ID of the Yandex Data Processing cluster.",
		Computed:    true,
		Optional:    true,
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
		client := dataprocsdk.NewClusterClient(config.SDK)

		clusterID, err = resolveObjectIDV2(ctx, config, d, func(name string, opts ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver {
			return dataprocsdk.ClusterResolver(name, client, opts...)
		})
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Data Processing Cluster by name: %v", err)
		}
	}

	client := dataprocsdk.NewClusterClient(config.SDK)

	cluster, err := client.Get(ctx, &dataproc.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", clusterID))
	}

	d.SetId(cluster.Id)
	return populateDataprocClusterResourceData(d, config, cluster)
}
