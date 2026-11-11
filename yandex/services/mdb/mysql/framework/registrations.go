package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/framework/mdb_mysql_cluster_v2"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/framework/mdb_mysql_database_v2"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/framework/mdb_mysql_user_v2"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_mysql_cluster_v2.NewMySQLClusterResourceV2,
		mdb_mysql_database_v2.NewResource,
		mdb_mysql_user_v2.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_mysql_database_v2.NewDataSource,
		mdb_mysql_user_v2.NewDataSource,
	}
}
