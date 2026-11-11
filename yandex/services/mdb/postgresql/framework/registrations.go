package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/framework/mdb_postgresql_cluster_v2"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_postgresql_cluster_v2.NewPostgreSQLClusterResourceV2,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
