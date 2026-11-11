package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/spark/framework/spark_cluster"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		spark_cluster.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		spark_cluster.NewDatasource,
	}
}
