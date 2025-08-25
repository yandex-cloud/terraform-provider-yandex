package mdb_sharded_postgresql_database

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
)

type Database struct {
	Id        types.String   `tfsdk:"id"`
	ClusterID types.String   `tfsdk:"cluster_id"`
	Name      types.String   `tfsdk:"name"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func dbToState(db *spqr.Database, state *Database) diag.Diagnostics {
	state.ClusterID = types.StringValue(db.ClusterId)
	state.Name = types.StringValue(db.Name)
	var diags diag.Diagnostics
	return diags
}

func dbFromState(state *Database) (*spqr.DatabaseSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	db := &spqr.DatabaseSpec{
		Name: state.Name.ValueString(),
	}

	return db, diags
}
