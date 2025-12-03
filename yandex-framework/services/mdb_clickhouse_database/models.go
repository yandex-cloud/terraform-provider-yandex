package mdb_clickhouse_database

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
)

type Database struct {
	Id        types.String   `tfsdk:"id"`
	ClusterID types.String   `tfsdk:"cluster_id"`
	Engine    types.String   `tfsdk:"engine"`
	Name      types.String   `tfsdk:"name"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func specToState(spec *clickhouse.Database, state *Database) {
	state.Id = types.StringValue(resourceid.Construct(spec.ClusterId, spec.Name))
	state.ClusterID = types.StringValue(spec.ClusterId)
	state.Name = types.StringValue(spec.Name)
	state.Engine = getDatabaseEngineName(spec.Engine)
}

func stateToSpec(state *Database, spec *clickhouse.DatabaseSpec) {
	spec.Name = state.Name.ValueString()
	dbEngine := getDatabaseEngineValue(state.Engine)
	if dbEngine != clickhouse.DatabaseEngine_DATABASE_ENGINE_UNSPECIFIED {
		spec.Engine = dbEngine
	}
}
