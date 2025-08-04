package mdb_sharded_postgresql_cluster

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type SettingsAttributeInfoProvider struct{}

func (p *SettingsAttributeInfoProvider) GetSettingsEnumNames() map[string]map[int32]string {
	return settingsEnumNames
}

func (p *SettingsAttributeInfoProvider) GetSettingsEnumValues() map[string]map[string]int32 {
	return settingsEnumValues
}

func (p *SettingsAttributeInfoProvider) GetSetAttributes() map[string]struct{} {
	return listAttributes
}

var settingsEnumNames = map[string]map[int32]string{
	"log_level":              config.LogLevel_name,
	"default_route_behavior": config.RouterSettings_DefaultRouteBehavior_name,
}

var settingsEnumValues = map[string]map[string]int32{
	"log_level":              config.LogLevel_value,
	"default_route_behavior": config.RouterSettings_DefaultRouteBehavior_value,
}

var listAttributes = map[string]struct{}{
	"default_route_behavior": {},
}

var attrProvider = &SettingsAttributeInfoProvider{}

func NewSettingsMapType() mdbcommon.SettingsMapType {
	return mdbcommon.NewSettingsMapType(attrProvider)
}

func NewSettingsMapValue(elements map[string]attr.Value) (mdbcommon.SettingsMapValue, diag.Diagnostics) {
	return mdbcommon.NewSettingsMapValue(elements, attrProvider)
}

func NewSettingsMapEmpty() mdbcommon.SettingsMapValue {
	s, _ := mdbcommon.NewSettingsMapValue(map[string]attr.Value{}, attrProvider)
	return s
}

func NewSettingsMapValueMust(elements map[string]attr.Value) mdbcommon.SettingsMapValue {
	val, d := NewSettingsMapValue(elements)
	if d.HasError() {
		panic(fmt.Sprintf("%v", d))
	}

	return val
}

func NewSettingsMapNull() mdbcommon.SettingsMapValue {
	return mdbcommon.NewSettingsMapNull()
}
