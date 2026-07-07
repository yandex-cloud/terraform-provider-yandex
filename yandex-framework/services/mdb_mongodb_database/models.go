package mdb_mongodb_database

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
)

type Database struct {
	Id                 types.String   `tfsdk:"id"`
	ClusterID          types.String   `tfsdk:"cluster_id"`
	Name               types.String   `tfsdk:"name"`
	DeletionProtection types.Bool     `tfsdk:"deletion_protection"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func databaseToState(db *mongodb.Database, state *Database) {
	state.ClusterID = types.StringValue(db.ClusterId)
	state.Name = types.StringValue(db.Name)
	state.DeletionProtection = wrappers.BoolToTF(db.GetDeletionProtection())
}
