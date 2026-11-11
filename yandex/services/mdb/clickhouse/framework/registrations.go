package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_database"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_user"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_clickhouse_cluster_v2.NewClickHouseClusterResourceV2,
		mdb_clickhouse_database.NewResource,
		mdb_clickhouse_user.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_clickhouse_cluster_v2.NewDataSource,
		mdb_clickhouse_database.NewDataSource,
		mdb_clickhouse_user.NewDataSource,
	}
}
