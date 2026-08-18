package usersettings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon"
)

var (
	UserSettings_OverflowMode_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "break",
	}
	UserSettings_OverflowMode_value     = chcommon.MakeReversedMap(UserSettings_OverflowMode_name, clickhouse.UserSettings_OverflowMode_value)
	UserSettings_OverflowMode_validator = chcommon.MakeEnumNamesValidator(UserSettings_OverflowMode_name)

	UserSettings_GroupByOverflowMode_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "break",
		3: "any",
	}
	UserSettings_GroupByOverflowMode_value     = chcommon.MakeReversedMap(UserSettings_GroupByOverflowMode_name, clickhouse.UserSettings_GroupByOverflowMode_value)
	UserSettings_GroupByOverflowMode_validator = chcommon.MakeEnumNamesValidator(UserSettings_GroupByOverflowMode_name)

	UserSettings_DistributedProductMode_name = map[int32]string{
		0: "unspecified",
		1: "deny",
		2: "local",
		3: "global",
		4: "allow",
	}
	UserSettings_DistributedProductMode_value     = chcommon.MakeReversedMap(UserSettings_DistributedProductMode_name, clickhouse.UserSettings_DistributedProductMode_value)
	UserSettings_DistributedProductMode_validator = chcommon.MakeEnumNamesValidator(UserSettings_DistributedProductMode_name)

	UserSettings_CountDistinctImplementation_name = map[int32]string{
		0: "unspecified",
		1: "uniq",
		2: "uniq_combined",
		3: "uniq_combined_64",
		4: "uniq_hll_12",
		5: "uniq_exact",
	}
	UserSettings_CountDistinctImplementation_value     = chcommon.MakeReversedMap(UserSettings_CountDistinctImplementation_name, clickhouse.UserSettings_CountDistinctImplementation_value)
	UserSettings_CountDistinctImplementation_validator = chcommon.MakeEnumNamesValidator(UserSettings_CountDistinctImplementation_name)

	UserSettings_QuotaMode_name = map[int32]string{
		0: "unspecified",
		1: "default",
		2: "keyed",
		3: "keyed_by_ip",
	}
	UserSettings_QuotaMode_value     = chcommon.MakeReversedMap(UserSettings_QuotaMode_name, clickhouse.UserSettings_QuotaMode_value)
	UserSettings_QuotaMode_validator = chcommon.MakeEnumNamesValidator(UserSettings_QuotaMode_name)

	UserSettings_LocalFilesystemReadMethod_name = map[int32]string{
		0: "unspecified",
		1: "read",
		2: "pread_threadpool",
		3: "pread",
		4: "nmap",
	}
	UserSettings_LocalFilesystemReadMethod_value     = chcommon.MakeReversedMap(UserSettings_LocalFilesystemReadMethod_name, clickhouse.UserSettings_LocalFilesystemReadMethod_value)
	UserSettings_LocalFilesystemReadMethod_validator = chcommon.MakeEnumNamesValidator(UserSettings_LocalFilesystemReadMethod_name)

	UserSettings_RemoteFilesystemReadMethod_name = map[int32]string{
		0: "unspecified",
		1: "read",
		2: "threadpool",
	}
	UserSettings_RemoteFilesystemReadMethod_value     = chcommon.MakeReversedMap(UserSettings_RemoteFilesystemReadMethod_name, clickhouse.UserSettings_RemoteFilesystemReadMethod_value)
	UserSettings_RemoteFilesystemReadMethod_validator = chcommon.MakeEnumNamesValidator(UserSettings_RemoteFilesystemReadMethod_name)

	UserSettings_LoadBalancing_name = map[int32]string{
		0: "unspecified",
		1: "random",
		2: "nearest_hostname",
		3: "in_order",
		4: "first_or_random",
		5: "round_robin",
	}
	UserSettings_LoadBalancing_value     = chcommon.MakeReversedMap(UserSettings_LoadBalancing_name, clickhouse.UserSettings_LoadBalancing_value)
	UserSettings_LoadBalancing_validator = chcommon.MakeEnumNamesValidator(UserSettings_LoadBalancing_name)

	UserSettings_DateTimeInputFormat_name = map[int32]string{
		0: "unspecified",
		1: "best_effort",
		2: "basic",
		3: "best_effort_us",
	}
	UserSettings_DateTimeInputFormat_value     = chcommon.MakeReversedMap(UserSettings_DateTimeInputFormat_name, clickhouse.UserSettings_DateTimeInputFormat_value)
	UserSettings_DateTimeInputFormat_validator = chcommon.MakeEnumNamesValidator(UserSettings_DateTimeInputFormat_name)

	UserSettings_DateTimeOutputFormat_name = map[int32]string{
		0: "unspecified",
		1: "simple",
		2: "iso",
		3: "unix_timestamp",
	}
	UserSettings_DateTimeOutputFormat_value     = chcommon.MakeReversedMap(UserSettings_DateTimeOutputFormat_name, clickhouse.UserSettings_DateTimeOutputFormat_value)
	UserSettings_DateTimeOutputFormat_validator = chcommon.MakeEnumNamesValidator(UserSettings_DateTimeOutputFormat_name)

	UserSettings_JoinAlgorithm_name = map[int32]string{
		0: "unspecified",
		1: "hash",
		2: "parallel_hash",
		3: "partial_merge",
		4: "direct",
		5: "auto",
		6: "full_sorting_merge",
		7: "prefer_partial_merge",
	}
	UserSettings_JoinAlgorithm_value     = chcommon.MakeReversedMap(UserSettings_JoinAlgorithm_name, clickhouse.UserSettings_JoinAlgorithm_value)
	UserSettings_JoinAlgorithm_validator = chcommon.MakeEnumNamesValidator(UserSettings_JoinAlgorithm_name)

	UserSettings_DistributedDdlOutputMode_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "none",
		3: "null_status_on_timeout",
		4: "never_throw",
		5: "none_only_active",
		6: "null_status_on_timeout_only_active",
		7: "throw_only_active",
	}
	UserSettings_DistributedDdlOutputMode_value     = chcommon.MakeReversedMap(UserSettings_DistributedDdlOutputMode_name, clickhouse.UserSettings_DistributedDdlOutputMode_value)
	UserSettings_DistributedDdlOutputMode_validator = chcommon.MakeEnumNamesValidator(UserSettings_DistributedDdlOutputMode_name)

	UserSettings_QueryCacheSystemTableHandling_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "save",
		3: "ignore",
	}
	UserSettings_QueryCacheSystemTableHandling_value     = chcommon.MakeReversedMap(UserSettings_QueryCacheSystemTableHandling_name, clickhouse.UserSettings_QueryCacheSystemTableHandling_value)
	UserSettings_QueryCacheSystemTableHandling_validator = chcommon.MakeEnumNamesValidator(UserSettings_QueryCacheSystemTableHandling_name)

	UserSettings_QueryCacheNondeterministicFunctionHandling_name = map[int32]string{
		0: "unspecified",
		1: "throw",
		2: "save",
		3: "ignore",
	}
	UserSettings_QueryCacheNondeterministicFunctionHandling_value     = chcommon.MakeReversedMap(UserSettings_QueryCacheNondeterministicFunctionHandling_name, clickhouse.UserSettings_QueryCacheNondeterministicFunctionHandling_value)
	UserSettings_QueryCacheNondeterministicFunctionHandling_validator = chcommon.MakeEnumNamesValidator(UserSettings_QueryCacheNondeterministicFunctionHandling_name)
)

func getOverflowModeName(value clickhouse.UserSettings_OverflowMode) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_OverflowMode_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getOverflowModeValue(name types.String) clickhouse.UserSettings_OverflowMode {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_OverflowMode_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_OverflowMode(value)
	}
	return 0
}

func getGroupByOverflowModeName(value clickhouse.UserSettings_GroupByOverflowMode) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_GroupByOverflowMode_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getGroupByOverflowModeValue(name types.String) clickhouse.UserSettings_GroupByOverflowMode {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_GroupByOverflowMode_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_GroupByOverflowMode(value)
	}
	return 0
}

func getDistributedProductModeName(value clickhouse.UserSettings_DistributedProductMode) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_DistributedProductMode_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getDistributedProductModeValue(name types.String) clickhouse.UserSettings_DistributedProductMode {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_DistributedProductMode_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_DistributedProductMode(value)
	}
	return 0
}

func getCountDistinctImplementationName(value clickhouse.UserSettings_CountDistinctImplementation) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_CountDistinctImplementation_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getCountDistinctImplementationValue(name types.String) clickhouse.UserSettings_CountDistinctImplementation {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_CountDistinctImplementation_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_CountDistinctImplementation(value)
	}
	return 0
}

func getQuotaModeName(value clickhouse.UserSettings_QuotaMode) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_QuotaMode_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getQuotaModeValue(name types.String) clickhouse.UserSettings_QuotaMode {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_QuotaMode_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_QuotaMode(value)
	}
	return 0
}

func getLocalFilesystemReadMethodName(value clickhouse.UserSettings_LocalFilesystemReadMethod) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_LocalFilesystemReadMethod_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getLocalFilesystemReadMethodValue(name types.String) clickhouse.UserSettings_LocalFilesystemReadMethod {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_LocalFilesystemReadMethod_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_LocalFilesystemReadMethod(value)
	}
	return 0
}

func getRemoteFilesystemReadMethodName(value clickhouse.UserSettings_RemoteFilesystemReadMethod) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_RemoteFilesystemReadMethod_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getRemoteFilesystemReadMethodValue(name types.String) clickhouse.UserSettings_RemoteFilesystemReadMethod {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_RemoteFilesystemReadMethod_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_RemoteFilesystemReadMethod(value)
	}
	return 0
}

func getLoadBalancingName(value clickhouse.UserSettings_LoadBalancing) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_LoadBalancing_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getLoadBalancingValue(name types.String) clickhouse.UserSettings_LoadBalancing {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_LoadBalancing_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_LoadBalancing(value)
	}
	return 0
}

func getDateTimeInputFormatName(value clickhouse.UserSettings_DateTimeInputFormat) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_DateTimeInputFormat_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getDateTimeInputFormatValue(name types.String) clickhouse.UserSettings_DateTimeInputFormat {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_DateTimeInputFormat_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_DateTimeInputFormat(value)
	}
	return 0
}

func getDateTimeOutputFormatName(value clickhouse.UserSettings_DateTimeOutputFormat) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_DateTimeOutputFormat_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getDateTimeOutputFormatValue(name types.String) clickhouse.UserSettings_DateTimeOutputFormat {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_DateTimeOutputFormat_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_DateTimeOutputFormat(value)
	}
	return 0
}

func getJoinAlgorithmName(value clickhouse.UserSettings_JoinAlgorithm) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_JoinAlgorithm_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getJoinAlgorithmValue(name types.String) clickhouse.UserSettings_JoinAlgorithm {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_JoinAlgorithm_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_JoinAlgorithm(value)
	}
	return 0
}

func getDistributedDdlOutputModeName(value clickhouse.UserSettings_DistributedDdlOutputMode) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_DistributedDdlOutputMode_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getDistributedDdlOutputModeValue(name types.String) clickhouse.UserSettings_DistributedDdlOutputMode {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_DistributedDdlOutputMode_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_DistributedDdlOutputMode(value)
	}
	return 0
}
func getQueryCacheSystemTableHandlingName(value clickhouse.UserSettings_QueryCacheSystemTableHandling) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_QueryCacheSystemTableHandling_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getQueryCacheSystemTableHandlingValue(name types.String) clickhouse.UserSettings_QueryCacheSystemTableHandling {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_QueryCacheSystemTableHandling_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_QueryCacheSystemTableHandling(value)
	}
	return 0
}
func getQueryCacheNondeterministicFunctionHandlingName(value clickhouse.UserSettings_QueryCacheNondeterministicFunctionHandling) types.String {
	if value == 0 {
		return types.StringNull()
	}
	if name, ok := UserSettings_QueryCacheNondeterministicFunctionHandling_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringUnknown()
}

func getQueryCacheNondeterministicFunctionHandlingValue(name types.String) clickhouse.UserSettings_QueryCacheNondeterministicFunctionHandling {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := UserSettings_QueryCacheNondeterministicFunctionHandling_value[name.ValueString()]; ok {
		return clickhouse.UserSettings_QueryCacheNondeterministicFunctionHandling(value)
	}
	return 0
}
