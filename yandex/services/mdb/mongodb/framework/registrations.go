package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mongodb/framework/mdb_mongodb_database"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mongodb/framework/mdb_mongodb_user"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		mdb_mongodb_database.NewResource,
		mdb_mongodb_user.NewResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		mdb_mongodb_database.NewDataSource,
		mdb_mongodb_user.NewDataSource,
	}
}
