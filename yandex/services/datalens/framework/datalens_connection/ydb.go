package datalens_connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

func flattenYdbToMap(ydb *ydbConfigModel, m map[string]interface{}) {
	if !ydb.WorkbookId.IsNull() && !ydb.WorkbookId.IsUnknown() {
		m["workbook_id"] = ydb.WorkbookId.ValueString()
	}
	if !ydb.DirPath.IsNull() && !ydb.DirPath.IsUnknown() {
		m["dir_path"] = ydb.DirPath.ValueString()
	}

	m["host"] = ydb.Host.ValueString()
	m["port"] = ydb.Port.ValueInt64()
	m["db_name"] = ydb.DbName.ValueString()
	m["cloud_id"] = ydb.CloudId.ValueString()
	m["folder_id"] = ydb.FolderId.ValueString()
	m["service_account_id"] = ydb.ServiceAccountId.ValueString()

	setStringIfKnown(m, "auth_type", ydb.AuthType)
	setStringIfKnown(m, "username", ydb.Username)
	setStringIfKnown(m, "token", ydb.Token)
	setStringIfKnown(m, "ssl_ca", ydb.SslCa)
	setStringIfKnown(m, "ssl_enable", ydb.SslEnable)
	setStringIfKnown(m, "raw_sql_level", ydb.RawSqlLevel)
	setStringIfKnown(m, "data_export_forbidden", ydb.DataExportForbidden)
	setStringIfKnown(m, "mdb_cluster_id", ydb.MdbClusterId)
	setStringIfKnown(m, "mdb_folder_id", ydb.MdbFolderId)

	if !ydb.CacheTtlSec.IsNull() && !ydb.CacheTtlSec.IsUnknown() {
		m["cache_ttl_sec"] = ydb.CacheTtlSec.ValueInt64()
	}

	if !ydb.DelegationIsSet.IsNull() && !ydb.DelegationIsSet.IsUnknown() {
		m["delegation_is_set"] = ydb.DelegationIsSet.ValueBool()
	}
}

// flattenYdbToUpdateMap builds the data map for updateConnection.
// It excludes immutable fields (workbook_id, dir_path) that are only set at creation time.
func flattenYdbToUpdateMap(ydb *ydbConfigModel, m map[string]interface{}) {
	m["host"] = ydb.Host.ValueString()
	m["port"] = ydb.Port.ValueInt64()
	m["db_name"] = ydb.DbName.ValueString()
	m["cloud_id"] = ydb.CloudId.ValueString()
	m["folder_id"] = ydb.FolderId.ValueString()
	m["service_account_id"] = ydb.ServiceAccountId.ValueString()

	setStringIfKnown(m, "auth_type", ydb.AuthType)
	setStringIfKnown(m, "username", ydb.Username)
	setStringIfKnown(m, "token", ydb.Token)
	setStringIfKnown(m, "ssl_ca", ydb.SslCa)
	setStringIfKnown(m, "ssl_enable", ydb.SslEnable)
	setStringIfKnown(m, "raw_sql_level", ydb.RawSqlLevel)
	setStringIfKnown(m, "data_export_forbidden", ydb.DataExportForbidden)
	setStringIfKnown(m, "mdb_cluster_id", ydb.MdbClusterId)
	setStringIfKnown(m, "mdb_folder_id", ydb.MdbFolderId)

	if !ydb.CacheTtlSec.IsNull() && !ydb.CacheTtlSec.IsUnknown() {
		m["cache_ttl_sec"] = ydb.CacheTtlSec.ValueInt64()
	}

	if !ydb.DelegationIsSet.IsNull() && !ydb.DelegationIsSet.IsUnknown() {
		m["delegation_is_set"] = ydb.DelegationIsSet.ValueBool()
	}
}

func populateYdbDataSourceFromResponse(ydb *ydbDataSourceConfigModel, resp map[string]interface{}) {
	if _, ok := resp["workbook_id"]; ok {
		ydb.WorkbookId = datalens.StringOrNull(resp, "workbook_id")
	}
	if _, ok := resp["dir_path"]; ok {
		ydb.DirPath = datalens.StringOrNull(resp, "dir_path")
	}

	if v, ok := resp["host"].(string); ok {
		ydb.Host = types.StringValue(v)
	}
	if v, ok := resp["port"]; ok {
		ydb.Port = types.Int64Value(datalens.ToInt64(v))
	}
	if v, ok := resp["db_name"].(string); ok {
		ydb.DbName = types.StringValue(v)
	}
	if v, ok := resp["cloud_id"].(string); ok {
		ydb.CloudId = types.StringValue(v)
	}
	if v, ok := resp["folder_id"].(string); ok {
		ydb.FolderId = types.StringValue(v)
	}
	if v, ok := resp["service_account_id"].(string); ok {
		ydb.ServiceAccountId = types.StringValue(v)
	}

	ydb.AuthType = datalens.StringOrNull(resp, "auth_type")
	ydb.Username = datalens.StringOrNull(resp, "username")
	ydb.SslEnable = datalens.StringOrNull(resp, "ssl_enable")
	ydb.RawSqlLevel = datalens.StringOrNull(resp, "raw_sql_level")
	ydb.DataExportForbidden = datalens.StringOrNull(resp, "data_export_forbidden")
	ydb.MdbClusterId = datalens.StringOrNull(resp, "mdb_cluster_id")
	ydb.MdbFolderId = datalens.StringOrNull(resp, "mdb_folder_id")

	if v, ok := resp["cache_ttl_sec"]; ok && v != nil {
		ydb.CacheTtlSec = types.Int64Value(datalens.ToInt64(v))
	} else {
		ydb.CacheTtlSec = types.Int64Null()
	}

	if v, ok := resp["delegation_is_set"]; ok && v != nil {
		if b, ok := v.(bool); ok {
			ydb.DelegationIsSet = types.BoolValue(b)
		} else {
			ydb.DelegationIsSet = types.BoolNull()
		}
	} else {
		ydb.DelegationIsSet = types.BoolNull()
	}
}

func populateYdbFromResponse(ydb *ydbConfigModel, resp map[string]interface{}) {
	// workbook_id and dir_path are returned by getConnection;
	// only overwrite if present in response to preserve existing state.
	if _, ok := resp["workbook_id"]; ok {
		ydb.WorkbookId = datalens.StringOrNull(resp, "workbook_id")
	}
	if _, ok := resp["dir_path"]; ok {
		ydb.DirPath = datalens.StringOrNull(resp, "dir_path")
	}

	if v, ok := resp["host"].(string); ok {
		ydb.Host = types.StringValue(v)
	}
	if v, ok := resp["port"]; ok {
		ydb.Port = types.Int64Value(datalens.ToInt64(v))
	}
	if v, ok := resp["db_name"].(string); ok {
		ydb.DbName = types.StringValue(v)
	}
	if v, ok := resp["cloud_id"].(string); ok {
		ydb.CloudId = types.StringValue(v)
	}
	if v, ok := resp["folder_id"].(string); ok {
		ydb.FolderId = types.StringValue(v)
	}
	if v, ok := resp["service_account_id"].(string); ok {
		ydb.ServiceAccountId = types.StringValue(v)
	}

	ydb.AuthType = datalens.StringOrNull(resp, "auth_type")
	ydb.Username = datalens.StringOrNull(resp, "username")
	ydb.SslEnable = datalens.StringOrNull(resp, "ssl_enable")
	ydb.RawSqlLevel = datalens.StringOrNull(resp, "raw_sql_level")
	ydb.DataExportForbidden = datalens.StringOrNull(resp, "data_export_forbidden")
	ydb.MdbClusterId = datalens.StringOrNull(resp, "mdb_cluster_id")
	ydb.MdbFolderId = datalens.StringOrNull(resp, "mdb_folder_id")

	if v, ok := resp["cache_ttl_sec"]; ok && v != nil {
		ydb.CacheTtlSec = types.Int64Value(datalens.ToInt64(v))
	} else {
		ydb.CacheTtlSec = types.Int64Null()
	}

	if v, ok := resp["delegation_is_set"]; ok && v != nil {
		if b, ok := v.(bool); ok {
			ydb.DelegationIsSet = types.BoolValue(b)
		} else {
			ydb.DelegationIsSet = types.BoolNull()
		}
	} else {
		ydb.DelegationIsSet = types.BoolNull()
	}
}
