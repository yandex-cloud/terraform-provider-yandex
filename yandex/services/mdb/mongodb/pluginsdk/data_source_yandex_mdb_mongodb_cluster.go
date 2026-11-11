package pluginsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	mongodbsdk "github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func DataSourceYandexMDBMongodbCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed MongoDB cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",

		ReadContext: dataSourceYandexMDBMongodbClusterRead,
		Schema:      sdkcommon.ConvertToOptional(ResourceYandexMDBMongodbCluster().Schema),
	}
}

func dataSourceYandexMDBMongodbClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// fix to import users and databases
	ctx = context.WithValue(ctx, ReadModeKey, true)
	config := meta.(*providerconfig.Config)

	err := sdkcommon.CheckOneOf(d, "cluster_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = sdkcommon.ResolveObjectIDV2(ctx, config, d,
			func(name string, opts ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver {
				return mongodbsdk.ClusterResolver(name, mongodbsdk.NewClusterClient(config.SDK), opts...)
			},
		)
		if err != nil {
			return diag.Errorf("failed to resolve data source Mongodb Cluster by name: %v", err)
		}

		if err := d.Set("cluster_id", clusterID); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(clusterID)

	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}
