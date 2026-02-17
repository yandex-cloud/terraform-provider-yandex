package datalens_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectionModel describes the resource data model.
// Common fields (id, type, name, description, created_at, updated_at, organization_id)
// live at the top level; connection-type-specific configuration is inside a nested attribute (e.g. ydb).
type connectionModel struct {
	Id             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	OrganizationId types.String `tfsdk:"organization_id"`

	// Connection-type-specific nested attributes (exactly one must be set)
	Ydb *ydbConfigModel `tfsdk:"ydb"`
}

type connectionDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	OrganizationId types.String `tfsdk:"organization_id"`

	Ydb *ydbDataSourceConfigModel `tfsdk:"ydb"`
}

type ydbConfigModel struct {
	WorkbookId types.String `tfsdk:"workbook_id"`
	DirPath    types.String `tfsdk:"dir_path"`

	// YDB-specific required fields
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	DbName           types.String `tfsdk:"db_name"`
	CloudId          types.String `tfsdk:"cloud_id"`
	FolderId         types.String `tfsdk:"folder_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`

	// YDB-specific optional fields
	AuthType            types.String `tfsdk:"auth_type"`
	Username            types.String `tfsdk:"username"`
	Token               types.String `tfsdk:"token"`
	SslCa               types.String `tfsdk:"ssl_ca"`
	SslEnable           types.String `tfsdk:"ssl_enable"`
	RawSqlLevel         types.String `tfsdk:"raw_sql_level"`
	CacheTtlSec         types.Int64  `tfsdk:"cache_ttl_sec"`
	DataExportForbidden types.String `tfsdk:"data_export_forbidden"`
	MdbClusterId        types.String `tfsdk:"mdb_cluster_id"`
	MdbFolderId         types.String `tfsdk:"mdb_folder_id"`
	DelegationIsSet     types.Bool   `tfsdk:"delegation_is_set"`
}

type ydbDataSourceConfigModel struct {
	WorkbookId types.String `tfsdk:"workbook_id"`
	DirPath    types.String `tfsdk:"dir_path"`

	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	DbName           types.String `tfsdk:"db_name"`
	CloudId          types.String `tfsdk:"cloud_id"`
	FolderId         types.String `tfsdk:"folder_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`

	AuthType            types.String `tfsdk:"auth_type"`
	Username            types.String `tfsdk:"username"`
	SslEnable           types.String `tfsdk:"ssl_enable"`
	RawSqlLevel         types.String `tfsdk:"raw_sql_level"`
	CacheTtlSec         types.Int64  `tfsdk:"cache_ttl_sec"`
	DataExportForbidden types.String `tfsdk:"data_export_forbidden"`
	MdbClusterId        types.String `tfsdk:"mdb_cluster_id"`
	MdbFolderId         types.String `tfsdk:"mdb_folder_id"`
	DelegationIsSet     types.Bool   `tfsdk:"delegation_is_set"`
}
