package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/gitlab/framework/gitlab_instance"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		gitlab_instance.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		gitlab_instance.NewDataSource,
	}
}
