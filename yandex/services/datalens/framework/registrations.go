package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datalens/framework/datalens_connection"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		datalens_connection.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datalens_connection.NewDataSource,
	}
}
