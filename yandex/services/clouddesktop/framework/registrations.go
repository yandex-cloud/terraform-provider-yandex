package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/clouddesktop/framework/cloud_desktops_desktop"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/clouddesktop/framework/cloud_desktops_desktop_group"
	cloud_desktops_image "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/clouddesktop/framework/cloud_desktops_image"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		cloud_desktops_desktop.NewResource,
		cloud_desktops_desktop_group.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloud_desktops_desktop.NewDatasource,
		cloud_desktops_desktop_group.NewDatasource,
		cloud_desktops_image.NewDataSource,
	}
}
