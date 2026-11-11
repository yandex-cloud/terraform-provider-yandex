package framework

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	airflow "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/airflow/framework"
	billing "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/billing/framework"
	cloud_desktop "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/clouddesktop/framework"
	cloud_registry "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cloudregistry/framework"
	datalens "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datalens/framework"
	datasphere "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/framework"
	managed_gitlab "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/gitlab/framework"
	managed_kubernetes "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/framework"
	clickhouse "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework"
	greenplum "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/framework"
	mongodb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mongodb/framework"
	mysql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/framework"
	opensearch "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/opensearch/framework"
	postgresql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/framework"
	redis "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/framework"
	sharded_postgresql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/framework"
	metastore "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/metastore/framework"
	organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/framework"
	spark "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/spark/framework"
	object_storage "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/framework"
	trino "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/trino/framework"
	vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/framework"
	yandex_query "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/yandex_query/framework"
	"slices"
)

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return slices.Concat(
		generatedResourcesWithCICLogging(ctx),
		airflow.Resources(),
		billing.Resources(),
		clickhouse.Resources(),
		cloud_desktop.Resources(),
		cloud_registry.Resources(),
		datalens.Resources(),
		datasphere.Resources(),
		greenplum.Resources(),
		managed_gitlab.Resources(),
		managed_kubernetes.Resources(),
		metastore.Resources(),
		mongodb.Resources(),
		mysql.Resources(),
		object_storage.Resources(),
		opensearch.Resources(),
		organization_manager.Resources(),
		postgresql.Resources(),
		redis.Resources(),
		sharded_postgresql.Resources(),
		spark.Resources(),
		trino.Resources(),
		vpc.Resources(),
		yandex_query.Resources(),
	)
}
func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return slices.Concat(
		generatedDataSourcesWithCICLogging(ctx),
		airflow.DataSources(),
		billing.DataSources(),
		clickhouse.DataSources(),
		cloud_desktop.DataSources(),
		cloud_registry.DataSources(),
		datalens.DataSources(),
		datasphere.DataSources(),
		greenplum.DataSources(),
		managed_gitlab.DataSources(),
		managed_kubernetes.DataSources(),
		metastore.DataSources(),
		mongodb.DataSources(),
		mysql.DataSources(),
		object_storage.DataSources(),
		opensearch.DataSources(),
		organization_manager.DataSources(),
		postgresql.DataSources(),
		redis.DataSources(),
		sharded_postgresql.DataSources(),
		spark.DataSources(),
		trino.DataSources(),
		vpc.DataSources(),
		yandex_query.DataSources(),
	)
}
