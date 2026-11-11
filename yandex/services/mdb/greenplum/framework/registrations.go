package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/framework/mdb_greenplum_cluster_v2"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/framework/mdb_greenplum_resource_group"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/framework/mdb_greenplum_user"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_greenplum_cluster_v2.NewResource,
		mdb_greenplum_resource_group.NewResource,
		mdb_greenplum_user.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_greenplum_cluster_v2.NewDataSource,
		mdb_greenplum_resource_group.NewDataSource,
		mdb_greenplum_user.NewDataSource,
	}
}
