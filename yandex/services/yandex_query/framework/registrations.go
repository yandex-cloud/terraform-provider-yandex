package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_monitoring_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_object_storage_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_object_storage_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_ydb_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_yds_binding"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework/yq_yds_connection"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		yq_monitoring_connection.NewResource,
		yq_object_storage_binding.NewResource,
		yq_object_storage_connection.NewResource,
		yq_ydb_connection.NewResource,
		yq_yds_binding.NewResource,
		yq_yds_connection.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return nil
}
