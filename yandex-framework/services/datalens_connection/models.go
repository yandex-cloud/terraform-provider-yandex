package datalens_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectionModel is both the Terraform-Framework model and the wire DTO
// for the DataLens connection REST API. The DataLens connection payload is
// flat (type/name/description side-by-side with type-specific fields like
// host/port/db_name). The type-specific block is `wire:"-"` here because
// (un)marshalConnection does an explicit two-pass: first the common parent
// fields, then the variant struct from the same flat response — keeping the
// nested HCL block (`ydb {}`) without `,inline` magic.
type connectionModel struct {
	Id             types.String `tfsdk:"id"              wire:"id"`
	Type           types.String `tfsdk:"type"            wire:"type"`
	Name           types.String `tfsdk:"name"            wire:"name"`
	Description    types.String `tfsdk:"description"     wire:"description,nullIfEmpty"`
	CreatedAt      types.String `tfsdk:"created_at"      wire:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"      wire:"updated_at"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`

	Ydb *ydbConfigModel `tfsdk:"ydb" wire:"-"`
}

type ydbConfigModel struct {
	WorkbookId types.String `tfsdk:"workbook_id" wire:"workbook_id"`
	DirPath    types.String `tfsdk:"dir_path"    wire:"dir_path"`

	Host             types.String `tfsdk:"host"               wire:"host"`
	Port             types.Int64  `tfsdk:"port"               wire:"port"`
	DbName           types.String `tfsdk:"db_name"            wire:"db_name"`
	CloudId          types.String `tfsdk:"cloud_id"           wire:"cloud_id"`
	FolderId         types.String `tfsdk:"folder_id"          wire:"folder_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id" wire:"service_account_id"`

	AuthType                               types.String `tfsdk:"auth_type"                                  wire:"auth_type"`
	Username                               types.String `tfsdk:"username"                                   wire:"username"`
	Token                                  types.String `tfsdk:"token"                                      wire:"token"`
	SslCa                                  types.String `tfsdk:"ssl_ca"                                     wire:"ssl_ca"`
	SslEnable                              types.String `tfsdk:"ssl_enable"                                 wire:"ssl_enable"`
	RawSqlLevel                            types.String `tfsdk:"raw_sql_level"                              wire:"raw_sql_level"`
	CacheTtlSec                            types.Int64  `tfsdk:"cache_ttl_sec"                              wire:"cache_ttl_sec"`
	CacheInvalidationThrottlingIntervalSec types.Int64  `tfsdk:"cache_invalidation_throttling_interval_sec" wire:"cache_invalidation_throttling_interval_sec"`
	DataExportForbidden                    types.String `tfsdk:"data_export_forbidden"                      wire:"data_export_forbidden"`
	MdbClusterId                           types.String `tfsdk:"mdb_cluster_id"                             wire:"mdb_cluster_id"`
	MdbFolderId                            types.String `tfsdk:"mdb_folder_id"                              wire:"mdb_folder_id"`
	DelegationIsSet                        types.Bool   `tfsdk:"delegation_is_set"                          wire:"delegation_is_set"`
}

type connectionDataSourceModel struct {
	Id             types.String `tfsdk:"id"              wire:"id"`
	Type           types.String `tfsdk:"type"            wire:"type"`
	Name           types.String `tfsdk:"name"            wire:"name"`
	Description    types.String `tfsdk:"description"     wire:"description,nullIfEmpty"`
	CreatedAt      types.String `tfsdk:"created_at"      wire:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"      wire:"updated_at"`
	OrganizationId types.String `tfsdk:"organization_id" wire:"-"`

	Ydb *ydbDataSourceConfigModel `tfsdk:"ydb" wire:"-"`
}

type ydbDataSourceConfigModel struct {
	WorkbookId types.String `tfsdk:"workbook_id" wire:"workbook_id"`
	DirPath    types.String `tfsdk:"dir_path"    wire:"dir_path"`

	Host             types.String `tfsdk:"host"               wire:"host"`
	Port             types.Int64  `tfsdk:"port"               wire:"port"`
	DbName           types.String `tfsdk:"db_name"            wire:"db_name"`
	CloudId          types.String `tfsdk:"cloud_id"           wire:"cloud_id"`
	FolderId         types.String `tfsdk:"folder_id"          wire:"folder_id"`
	ServiceAccountId types.String `tfsdk:"service_account_id" wire:"service_account_id"`

	AuthType                               types.String `tfsdk:"auth_type"                                  wire:"auth_type"`
	Username                               types.String `tfsdk:"username"                                   wire:"username"`
	SslEnable                              types.String `tfsdk:"ssl_enable"                                 wire:"ssl_enable"`
	RawSqlLevel                            types.String `tfsdk:"raw_sql_level"                              wire:"raw_sql_level"`
	CacheTtlSec                            types.Int64  `tfsdk:"cache_ttl_sec"                              wire:"cache_ttl_sec"`
	CacheInvalidationThrottlingIntervalSec types.Int64  `tfsdk:"cache_invalidation_throttling_interval_sec" wire:"cache_invalidation_throttling_interval_sec"`
	DataExportForbidden                    types.String `tfsdk:"data_export_forbidden"                      wire:"data_export_forbidden"`
	MdbClusterId                           types.String `tfsdk:"mdb_cluster_id"                             wire:"mdb_cluster_id"`
	MdbFolderId                            types.String `tfsdk:"mdb_folder_id"                              wire:"mdb_folder_id"`
	DelegationIsSet                        types.Bool   `tfsdk:"delegation_is_set"                          wire:"delegation_is_set"`
}
