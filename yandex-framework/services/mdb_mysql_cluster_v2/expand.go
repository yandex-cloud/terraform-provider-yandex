package mdb_mysql_cluster_v2

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

// Set access to default if null
func expandAccess(ctx context.Context, cfgAccess types.Object, diags *diag.Diagnostics) *mysql.Access {
	var access Access
	diags.Append(cfgAccess.As(ctx, &access, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}
	return &mysql.Access{
		WebSql:       access.WebSql.ValueBool(),
		DataLens:     access.DataLens.ValueBool(),
		DataTransfer: access.DataTransfer.ValueBool(),
	}
}

const (
	anytimeType = "ANYTIME" //nolint:unused
	weeklyType  = "WEEKLY"  //nolint:unused
)

func expandPerformanceDiagnostics(ctx context.Context, pd types.Object, diags *diag.Diagnostics) *mysql.PerformanceDiagnostics {
	if pd.IsNull() || pd.IsUnknown() {
		return nil
	}
	var pdConf PerformanceDiagnostics

	diags.Append(pd.As(ctx, &pdConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &mysql.PerformanceDiagnostics{
		Enabled:                    pdConf.Enabled.ValueBool(),
		SessionsSamplingInterval:   pdConf.SessionsSamplingInterval.ValueInt64(),
		StatementsSamplingInterval: pdConf.StatementsSamplingInterval.ValueInt64(),
	}
}

var msVersionConfig = map[string]mysql.ConfigSpec_MysqlConfig{
	"5.7": &mysql.ConfigSpec_MysqlConfig_5_7{},
	"8.0": &mysql.ConfigSpec_MysqlConfig_8_0{},
}

func expandMySQLConfig(
	ctx context.Context,
	version string, config mdbcommon.SettingsMapValue,
	diags *diag.Diagnostics,
) mysql.ConfigSpec_MysqlConfig {

	a := protobuf_adapter.NewProtobufMapDataAdapter()

	if msVersionConfig[version] == nil {
		diags.AddError("Failed to expand MySQL config.", fmt.Sprintf("unsupported version %s.", version))
		return nil
	}

	msConf := reflect.New(reflect.TypeOf(msVersionConfig[version]).Elem()).Interface()
	if diags.HasError() {
		return nil
	}

	attrs := config.PrimitiveElements(ctx, diags)
	a.Fill(ctx, msConf, attrs, diags)

	return msConf.(mysql.ConfigSpec_MysqlConfig)
}

func expandConfig(ctx context.Context, configSpec Config, diags *diag.Diagnostics) *mysql.ConfigSpec {
	return &mysql.ConfigSpec{
		Version:                configSpec.Version.ValueString(),
		Resources:              mdbcommon.ExpandResources[mysql.Resources](ctx, configSpec.Resources, diags),
		Access:                 expandAccess(ctx, configSpec.Access, diags),
		PerformanceDiagnostics: expandPerformanceDiagnostics(ctx, configSpec.PerformanceDiagnostics, diags),
		BackupRetainPeriodDays: mdbcommon.ExpandInt64Wrapper(ctx, configSpec.BackupRetainPeriodDays, diags),
		BackupWindowStart:      mdbcommon.ExpandBackupWindow(ctx, configSpec.BackupWindowStart, diags),
		MysqlConfig:            expandMySQLConfig(ctx, configSpec.Version.ValueString(), configSpec.MySQLConfig, diags),
	}
}
