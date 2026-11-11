package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/trino/framework/trino_access_control"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/trino/framework/trino_catalog"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/trino/framework/trino_cluster"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		trino_access_control.NewResource,
		trino_catalog.NewResource,
		trino_cluster.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		trino_access_control.NewDatasource,
		trino_catalog.NewDatasource,
		trino_cluster.NewDatasource,
	}
}
