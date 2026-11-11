package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/framework/mdb_redis_cluster_v2"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/framework/mdb_redis_user"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_redis_cluster_v2.NewResource,
		mdb_redis_user.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_redis_cluster_v2.NewDataSource,
		mdb_redis_user.NewDataSource,
	}
}
