package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cloudregistry/framework/cloudregistry_registry_ip_permission"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		cloudregistry_registry_ip_permission.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloudregistry_registry_ip_permission.NewDataSource,
	}
}
