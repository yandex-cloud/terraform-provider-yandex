package mdb_mysql_cluster_beta

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type MsSettingsAttributeInfoProvider struct{}

func (p *MsSettingsAttributeInfoProvider) GetSettingsEnumNames() map[string]map[int32]string {
	return msSettingsEnumNames
}

func (p *MsSettingsAttributeInfoProvider) GetSettingsEnumValues() map[string]map[string]int32 {
	return msSettingsEnumValues
}

func (p *MsSettingsAttributeInfoProvider) GetSetAttributes() map[string]struct{} {
	return listAttributes
}

var msSettingsEnumNames = map[string]map[int32]string{
	"default_authentication_plugin": config.MysqlConfig8_0_AuthPlugin_name,
	"transaction_isolation":         config.MysqlConfig8_0_TransactionIsolation_name,
	"binlog_row_image":              config.MysqlConfig8_0_BinlogRowImage_name,
	"slave_parallel_type":           config.MysqlConfig8_0_SlaveParallelType_name,
	"sql_mode.element":              config.MysqlConfig8_0_SQLMode_name,
}

var msSettingsEnumValues = map[string]map[string]int32{
	"default_authentication_plugin": config.MysqlConfig8_0_AuthPlugin_value,
	"transaction_isolation":         config.MysqlConfig8_0_TransactionIsolation_value,
	"binlog_row_image":              config.MysqlConfig8_0_BinlogRowImage_value,
	"slave_parallel_type":           config.MysqlConfig8_0_SlaveParallelType_value,
	"sql_mode.element":              config.MysqlConfig8_0_SQLMode_value,
}

var listAttributes = map[string]struct{}{
	"sql_mode": {},
}

var msAttrProvider = &MsSettingsAttributeInfoProvider{}

func NewMsSettingsMapType() mdbcommon.SettingsMapType {
	return mdbcommon.NewSettingsMapType(msAttrProvider)
}

func NewMsSettingsMapValue(elements map[string]attr.Value) (mdbcommon.SettingsMapValue, diag.Diagnostics) {
	return mdbcommon.NewSettingsMapValue(elements, msAttrProvider)
}

func NewMsSettingsMapValueMust(elements map[string]attr.Value) mdbcommon.SettingsMapValue {
	val, d := NewMsSettingsMapValue(elements)
	if d.HasError() {
		panic(fmt.Sprintf("%v", d))
	}

	return val
}

func NewMsSettingsMapNull() mdbcommon.SettingsMapValue {
	return mdbcommon.NewSettingsMapNull()
}
