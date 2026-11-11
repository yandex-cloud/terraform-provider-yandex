package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/framework/mdb_sharded_postgresql_cluster"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/framework/mdb_sharded_postgresql_database"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/framework/mdb_sharded_postgresql_shard"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/framework/mdb_sharded_postgresql_user"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_sharded_postgresql_cluster.NewShardedPostgreSQLClusterResource,
		mdb_sharded_postgresql_database.NewShardedPostgreSQLDatabaseResource,
		mdb_sharded_postgresql_shard.NewShardedPostgreSQLShardResource,
		mdb_sharded_postgresql_user.NewShardedPostgreSQLUserResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
