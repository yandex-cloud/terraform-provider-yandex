package mdb_postgresql_cluster_v2

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

var pgSettingsEnumNames = map[string]map[int32]string{
	"wal_level":                        config.PostgresqlConfig13_WalLevel_name,
	"synchronous_commit":               config.PostgresqlConfig13_SynchronousCommit_name,
	"constraint_exclusion":             config.PostgresqlConfig13_ConstraintExclusion_name,
	"force_parallel_mode":              config.PostgresqlConfig13_ForceParallelMode_name,
	"client_min_messages":              config.PostgresqlConfig13_LogLevel_name,
	"log_min_messages":                 config.PostgresqlConfig13_LogLevel_name,
	"log_min_error_statement":          config.PostgresqlConfig13_LogLevel_name,
	"log_error_verbosity":              config.PostgresqlConfig13_LogErrorVerbosity_name,
	"log_statement":                    config.PostgresqlConfig13_LogStatement_name,
	"default_transaction_isolation":    config.PostgresqlConfig13_TransactionIsolation_name,
	"bytea_output":                     config.PostgresqlConfig13_ByteaOutput_name,
	"xmlbinary":                        config.PostgresqlConfig13_XmlBinary_name,
	"xmloption":                        config.PostgresqlConfig13_XmlOption_name,
	"backslash_quote":                  config.PostgresqlConfig13_BackslashQuote_name,
	"plan_cache_mode":                  config.PostgresqlConfig13_PlanCacheMode_name,
	"pg_hint_plan_debug_print":         config.PostgresqlConfig13_PgHintPlanDebugPrint_name,
	"pg_hint_plan_message_level":       config.PostgresqlConfig13_LogLevel_name,
	"shared_preload_libraries.element": config.PostgresqlConfig13_SharedPreloadLibraries_name,
	"password_encryption":              config.PostgresqlConfig13_PasswordEncryption_name,
}

var pgSettingsEnumValues = map[string]map[string]int32{
	"wal_level":                        config.PostgresqlConfig13_WalLevel_value,
	"synchronous_commit":               config.PostgresqlConfig13_SynchronousCommit_value,
	"constraint_exclusion":             config.PostgresqlConfig13_ConstraintExclusion_value,
	"force_parallel_mode":              config.PostgresqlConfig13_ForceParallelMode_value,
	"client_min_messages":              config.PostgresqlConfig13_LogLevel_value,
	"log_min_messages":                 config.PostgresqlConfig13_LogLevel_value,
	"log_min_error_statement":          config.PostgresqlConfig13_LogLevel_value,
	"log_error_verbosity":              config.PostgresqlConfig13_LogErrorVerbosity_value,
	"log_statement":                    config.PostgresqlConfig13_LogStatement_value,
	"default_transaction_isolation":    config.PostgresqlConfig13_TransactionIsolation_value,
	"bytea_output":                     config.PostgresqlConfig13_ByteaOutput_value,
	"xmlbinary":                        config.PostgresqlConfig13_XmlBinary_value,
	"xmloption":                        config.PostgresqlConfig13_XmlOption_value,
	"backslash_quote":                  config.PostgresqlConfig13_BackslashQuote_value,
	"plan_cache_mode":                  config.PostgresqlConfig13_PlanCacheMode_value,
	"pg_hint_plan_debug_print":         config.PostgresqlConfig13_PgHintPlanDebugPrint_value,
	"pg_hint_plan_message_level":       config.PostgresqlConfig13_LogLevel_value,
	"shared_preload_libraries.element": config.PostgresqlConfig13_SharedPreloadLibraries_value,
	"password_encryption":              config.PostgresqlConfig13_PasswordEncryption_value,
}

var listAttributes = map[string]struct{}{
	"shared_preload_libraries": {},
}

var pgAttrProvider = &PgSettingsAttributeInfoProvider{}

type PgSettingsAttributeInfoProvider struct{}

func (p *PgSettingsAttributeInfoProvider) GetSettingsEnumNames() map[string]map[int32]string {
	return pgSettingsEnumNames
}

func (p *PgSettingsAttributeInfoProvider) GetSettingsEnumValues() map[string]map[string]int32 {
	return pgSettingsEnumValues
}

func (p *PgSettingsAttributeInfoProvider) GetSetAttributes() map[string]struct{} {
	return listAttributes
}

func NewPgSettingsMapType() mdbcommon.SettingsMapType {
	return mdbcommon.NewSettingsMapType(pgAttrProvider)
}

func NewPgSettingsMapValue(elements map[string]attr.Value) (mdbcommon.SettingsMapValue, diag.Diagnostics) {
	return mdbcommon.NewSettingsMapValue(elements, pgAttrProvider)
}

func NewPgSettingsMapValueMust(elements map[string]attr.Value) mdbcommon.SettingsMapValue {
	val, d := NewPgSettingsMapValue(elements)
	if d.HasError() {
		panic(fmt.Sprintf("%v", d))
	}

	return val
}

func NewPgSettingsMapNull() mdbcommon.SettingsMapValue {
	return mdbcommon.NewSettingsMapNull()
}
