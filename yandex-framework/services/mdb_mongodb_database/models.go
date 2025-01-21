package mdb_mongodb_database

import "github.com/hashicorp/terraform-plugin-framework/types"

type Database struct {
	Id        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`
}
