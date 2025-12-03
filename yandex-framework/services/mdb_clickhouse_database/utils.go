package mdb_clickhouse_database

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

var (
	Database_Engine_name = map[int32]string{
		0: "unspecified",
		1: "atomic",
		2: "replicated",
	}
	Database_Engine_value     = makeReversedMap(Database_Engine_name, clickhouse.DatabaseEngine_value)
	Database_Engine_validator = makeEnumNamesValidator(Database_Engine_name)
)

func getDatabaseEngineName(value clickhouse.DatabaseEngine) types.String {
	if name, ok := Database_Engine_name[int32(value)]; ok {
		return types.StringValue(name)
	}
	return types.StringValue(Database_Engine_name[0])
}

func getDatabaseEngineValue(name types.String) clickhouse.DatabaseEngine {
	if name.IsNull() || name.IsUnknown() {
		return 0
	}
	if value, ok := Database_Engine_value[name.ValueString()]; ok {
		return clickhouse.DatabaseEngine(value)
	}
	return 0
}

func makeReversedMap(m map[int32]string, addMap map[string]int32) map[string]int32 {
	r := addMap
	for k, v := range m {
		r[v] = k
	}
	return r
}

func makeEnumNamesValidator(m map[int32]string) []validator.String {
	res := make([]string, 0, len(m))
	for _, val := range m {
		res = append(res, val)
	}
	return []validator.String{stringvalidator.OneOf(res...)}
}
