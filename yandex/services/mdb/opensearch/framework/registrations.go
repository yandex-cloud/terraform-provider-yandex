package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/opensearch/framework/mdb_opensearch_cluster"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_opensearch_cluster.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_opensearch_cluster.NewDataSource,
	}
}
