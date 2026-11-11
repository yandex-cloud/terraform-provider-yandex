package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/billing/framework/billing_cloud_binding"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		billing_cloud_binding.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		billing_cloud_binding.NewDataSource,
	}
}
