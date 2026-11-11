package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/airflow/framework/airflow_cluster"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		airflow_cluster.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		airflow_cluster.NewDatasource,
	}
}
