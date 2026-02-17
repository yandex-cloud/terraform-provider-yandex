package datalens_connection

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenYdbToMap_RequiredFields(t *testing.T) {
	t.Parallel()

	ydb := &ydbConfigModel{
		Host:                types.StringValue("ydb.example.com"),
		Port:                types.Int64Value(2135),
		DbName:              types.StringValue("/ru-central1/b1g123/etn456"),
		CloudId:             types.StringValue("b1g123"),
		FolderId:            types.StringValue("b1g456"),
		ServiceAccountId:    types.StringValue("aje789"),
		WorkbookId:          types.StringNull(),
		DirPath:             types.StringNull(),
		AuthType:            types.StringNull(),
		Username:            types.StringNull(),
		Token:               types.StringNull(),
		SslCa:               types.StringNull(),
		SslEnable:           types.StringNull(),
		RawSqlLevel:         types.StringNull(),
		CacheTtlSec:         types.Int64Null(),
		DataExportForbidden: types.StringNull(),
		MdbClusterId:        types.StringNull(),
		MdbFolderId:         types.StringNull(),
		DelegationIsSet:     types.BoolNull(),
	}

	m := make(map[string]interface{})
	flattenYdbToMap(ydb, m)

	expectedRequired := map[string]interface{}{
		"host":               "ydb.example.com",
		"port":               int64(2135),
		"db_name":            "/ru-central1/b1g123/etn456",
		"cloud_id":           "b1g123",
		"folder_id":          "b1g456",
		"service_account_id": "aje789",
	}

	for k, v := range expectedRequired {
		if m[k] != v {
			t.Errorf("field %q: got %v, want %v", k, m[k], v)
		}
	}

	optionalKeys := []string{
		"workbook_id", "dir_path", "auth_type", "username", "token",
		"ssl_ca", "ssl_enable", "raw_sql_level", "cache_ttl_sec",
		"data_export_forbidden", "mdb_cluster_id", "mdb_folder_id", "delegation_is_set",
	}
	for _, k := range optionalKeys {
		if _, ok := m[k]; ok {
			t.Errorf("optional field %q should not be in map when null, but got %v", k, m[k])
		}
	}
}

func TestFlattenYdbToMap_AllFields(t *testing.T) {
	t.Parallel()

	ydb := &ydbConfigModel{
		WorkbookId:          types.StringValue("wb-123"),
		DirPath:             types.StringNull(), // null when workbook is set
		Host:                types.StringValue("ydb.example.com"),
		Port:                types.Int64Value(2135),
		DbName:              types.StringValue("/ru-central1/b1g123/etn456"),
		CloudId:             types.StringValue("b1g123"),
		FolderId:            types.StringValue("b1g456"),
		ServiceAccountId:    types.StringValue("aje789"),
		AuthType:            types.StringValue("password"),
		Username:            types.StringValue("admin"),
		Token:               types.StringNull(),
		SslCa:               types.StringValue("-----BEGIN CERTIFICATE-----"),
		SslEnable:           types.StringValue("on"),
		RawSqlLevel:         types.StringValue("subselect"),
		CacheTtlSec:         types.Int64Value(300),
		DataExportForbidden: types.StringValue("off"),
		MdbClusterId:        types.StringValue("mdb-cluster-1"),
		MdbFolderId:         types.StringValue("mdb-folder-1"),
		DelegationIsSet:     types.BoolValue(true),
	}

	m := make(map[string]interface{})
	flattenYdbToMap(ydb, m)

	expected := map[string]interface{}{
		"workbook_id":           "wb-123",
		"host":                  "ydb.example.com",
		"port":                  int64(2135),
		"db_name":               "/ru-central1/b1g123/etn456",
		"cloud_id":              "b1g123",
		"folder_id":             "b1g456",
		"service_account_id":    "aje789",
		"auth_type":             "password",
		"username":              "admin",
		"ssl_ca":                "-----BEGIN CERTIFICATE-----",
		"ssl_enable":            "on",
		"raw_sql_level":         "subselect",
		"cache_ttl_sec":         int64(300),
		"data_export_forbidden": "off",
		"mdb_cluster_id":        "mdb-cluster-1",
		"mdb_folder_id":         "mdb-folder-1",
		"delegation_is_set":     true,
	}

	if !reflect.DeepEqual(m, expected) {
		t.Errorf("flattenYdbToMap with all fields:\n  got  %v\n  want %v", m, expected)
	}

	if _, ok := m["dir_path"]; ok {
		t.Error("dir_path should not be in map when null")
	}
	if _, ok := m["token"]; ok {
		t.Error("token should not be in map when null")
	}
}

func TestPopulateYdbFromResponse_FullResponse(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"host":                  "ydb.example.com",
		"port":                  float64(2135), // JSON numbers are float64
		"db_name":               "/ru-central1/b1g123/etn456",
		"cloud_id":              "b1g123",
		"folder_id":             "b1g456",
		"service_account_id":    "aje789",
		"auth_type":             "password",
		"username":              "admin",
		"ssl_enable":            "on",
		"raw_sql_level":         "subselect",
		"cache_ttl_sec":         float64(300),
		"data_export_forbidden": "off",
		"mdb_cluster_id":        "mdb-cluster-1",
		"mdb_folder_id":         "mdb-folder-1",
		"delegation_is_set":     true,
	}

	ydb := &ydbConfigModel{}
	populateYdbFromResponse(ydb, resp)

	assertEqual(t, "host", ydb.Host, types.StringValue("ydb.example.com"))
	assertEqual(t, "port", ydb.Port, types.Int64Value(2135))
	assertEqual(t, "db_name", ydb.DbName, types.StringValue("/ru-central1/b1g123/etn456"))
	assertEqual(t, "cloud_id", ydb.CloudId, types.StringValue("b1g123"))
	assertEqual(t, "folder_id", ydb.FolderId, types.StringValue("b1g456"))
	assertEqual(t, "service_account_id", ydb.ServiceAccountId, types.StringValue("aje789"))
	assertEqual(t, "auth_type", ydb.AuthType, types.StringValue("password"))
	assertEqual(t, "username", ydb.Username, types.StringValue("admin"))
	assertEqual(t, "ssl_enable", ydb.SslEnable, types.StringValue("on"))
	assertEqual(t, "raw_sql_level", ydb.RawSqlLevel, types.StringValue("subselect"))
	assertEqual(t, "cache_ttl_sec", ydb.CacheTtlSec, types.Int64Value(300))
	assertEqual(t, "data_export_forbidden", ydb.DataExportForbidden, types.StringValue("off"))
	assertEqual(t, "mdb_cluster_id", ydb.MdbClusterId, types.StringValue("mdb-cluster-1"))
	assertEqual(t, "mdb_folder_id", ydb.MdbFolderId, types.StringValue("mdb-folder-1"))
	assertEqual(t, "delegation_is_set", ydb.DelegationIsSet, types.BoolValue(true))
}

func TestPopulateYdbFromResponse_MinimalResponse(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"host":               "ydb.example.com",
		"port":               float64(2135),
		"db_name":            "/path/to/db",
		"cloud_id":           "cloud-1",
		"folder_id":          "folder-1",
		"service_account_id": "sa-1",
	}

	ydb := &ydbConfigModel{}
	populateYdbFromResponse(ydb, resp)

	assertEqual(t, "host", ydb.Host, types.StringValue("ydb.example.com"))
	assertEqual(t, "port", ydb.Port, types.Int64Value(2135))
	assertEqual(t, "auth_type", ydb.AuthType, types.StringNull())
	assertEqual(t, "username", ydb.Username, types.StringNull())
	assertEqual(t, "ssl_enable", ydb.SslEnable, types.StringNull())
	assertEqual(t, "raw_sql_level", ydb.RawSqlLevel, types.StringNull())
	assertEqual(t, "cache_ttl_sec", ydb.CacheTtlSec, types.Int64Null())
	assertEqual(t, "data_export_forbidden", ydb.DataExportForbidden, types.StringNull())
	assertEqual(t, "mdb_cluster_id", ydb.MdbClusterId, types.StringNull())
	assertEqual(t, "mdb_folder_id", ydb.MdbFolderId, types.StringNull())
	assertEqual(t, "delegation_is_set", ydb.DelegationIsSet, types.BoolNull())
}

func TestPopulateYdbFromResponse_NullOptionalFields(t *testing.T) {
	t.Parallel()

	resp := map[string]interface{}{
		"host":               "ydb.example.com",
		"port":               float64(2135),
		"db_name":            "/path/to/db",
		"cloud_id":           "cloud-1",
		"folder_id":          "folder-1",
		"service_account_id": "sa-1",
		"auth_type":          nil,
		"cache_ttl_sec":      nil,
		"delegation_is_set":  nil,
	}

	ydb := &ydbConfigModel{}
	populateYdbFromResponse(ydb, resp)

	assertEqual(t, "auth_type", ydb.AuthType, types.StringNull())
	assertEqual(t, "cache_ttl_sec", ydb.CacheTtlSec, types.Int64Null())
	assertEqual(t, "delegation_is_set", ydb.DelegationIsSet, types.BoolNull())
}
