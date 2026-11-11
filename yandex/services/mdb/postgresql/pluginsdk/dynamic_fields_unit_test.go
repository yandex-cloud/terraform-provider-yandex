package pluginsdk_test

import (
	"testing"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	svc_postgresql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/pluginsdk"
)

func TestFieldsDynamicGenerateMapSchemaValidateFuncCorrect(t *testing.T) {
	t.Parallel()

	validateFunc := sdkcommon.GenerateMapSchemaValidateFunc(svc_postgresql.MdbPGSettingsFieldsInfo17)

	value := map[string]interface{}{
		"autovacuum_vacuum_scale_factor": "0.32",
		"default_transaction_isolation":  "TRANSACTION_ISOLATION_READ_UNCOMMITTED",
		"enable_parallel_hash":           "true",
		"max_connections":                "395",
		"shared_preload_libraries":       "SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_HINT_PLAN,SHARED_PRELOAD_LIBRARIES_ANON",
		"auto_explain_log_format":        "AUTO_EXPLAIN_LOG_FORMAT_XML",
	}

	_, errors := validateFunc(value, "")

	if len(errors) > 0 {
		for _, err := range errors {
			t.Errorf("generateMapSchemaValidateFunc: Error on correct example: %v", err)
		}
	}

}

func TestFieldsDynamicGenerateMapSchemaValidateFuncFail(t *testing.T) {
	t.Parallel()

	validateFunc := sdkcommon.GenerateMapSchemaValidateFunc(svc_postgresql.MdbPGSettingsFieldsInfo18)

	value := map[string]interface{}{
		// type is not correct string but should be int
		"autovacuum_vacuum_scale_factor": "a0.32",
		// Value not in enum
		"default_transaction_isolation": "TRANSACTION_OLATION_READ_UNCOMMITTED",
		// type is not correct int but should be bool
		"enable_parallel_hash": "5",
		// type is not correct bool but should be int
		"max_connections": "true",
		// type is not correct string but should be float
		"vacuum_cleanup_index_scale_factor": "p0.2",
	}

	_, errors := validateFunc(value, "")

	if len(errors) != 5 {
		t.Errorf("generateMapSchemaValidateFunc: Error on not correct example should be 5 errors, but out only: %v", len(errors))
		for _, err := range errors {
			t.Errorf("generateMapSchemaValidateFunc: Error on not correct example was: %v", err)
		}
	}

}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncFloat(t *testing.T) {
	t.Parallel()

	suppressFunc := sdkcommon.GenerateMapSchemaDiffSuppressFunc(svc_postgresql.MdbPGSettingsFieldsInfo17)

	if suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "0.33", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be not equal (0.32 != 0.33)")
	}

	if !suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "0.32", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be equal (0.32 == 0.32)")
	}

	if !suppressFunc("ttt.autovacuum_vacuum_scale_factor", "0.32", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: float values should be equal when new value is empty")
	}
}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncInt(t *testing.T) {
	t.Parallel()

	suppressFunc := sdkcommon.GenerateMapSchemaDiffSuppressFunc(svc_postgresql.MdbPGSettingsFieldsInfo17)

	if suppressFunc("ttt.max_connections", "32", "33", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be not equal (32 != 33)")
	}

	if !suppressFunc("ttt.max_connections", "32", "32", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal (32 == 32)")
	}

	if !suppressFunc("ttt.max_connections", "32", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal when new value is empty")
	}
}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncBool(t *testing.T) {
	t.Parallel()

	suppressFunc := sdkcommon.GenerateMapSchemaDiffSuppressFunc(svc_postgresql.MdbPGSettingsFieldsInfo17)

	if suppressFunc("ttt.enable_parallel_hash", "true", "false", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be not equal (true != false)")
	}

	if !suppressFunc("ttt.enable_parallel_hash", "true", "true", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal (true == true)")
	}

	if !suppressFunc("ttt.enable_parallel_hash", "true", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: int values should be equal when new value is empty")
	}
}

func TestFieldsDynamicGenerateMapSchemaDiffSuppressFuncEnum(t *testing.T) {
	t.Parallel()

	suppressFunc := sdkcommon.GenerateMapSchemaDiffSuppressFunc(svc_postgresql.MdbPGUserSettingsFieldsInfo)

	if suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "read committed", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be not equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED != read committed)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "read uncommitted", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED == read uncommitted)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal (TRANSACTION_ISOLATION_READ_UNCOMMITTED == TRANSACTION_ISOLATION_READ_UNCOMMITTED)")
	}

	if !suppressFunc("ttt.default_transaction_isolation", "TRANSACTION_ISOLATION_READ_UNCOMMITTED", "", nil) {
		t.Errorf("generateMapSchemaDiffSuppressFunc: enum values should be equal when new value is empty")
	}
}
