package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/framework/datasphere_community"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/framework/datasphere_project"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		datasphere_community.NewResource,
		datasphere_project.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasphere_community.NewDataSource,
		datasphere_project.NewDataSource,
	}
}
