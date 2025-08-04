package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Set access to default if null
func expandAccess(ctx context.Context, cfgAccess types.Object, diags *diag.Diagnostics) *postgresql.Access {
	var access Access
	diags.Append(cfgAccess.As(ctx, &access, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}
	return &postgresql.Access{
		WebSql:       access.WebSql.ValueBool(),
		DataLens:     access.DataLens.ValueBool(),
		DataTransfer: access.DataTransfer.ValueBool(),
		Serverless:   access.Serverless.ValueBool(),
	}
}

const (
	anytimeType = "ANYTIME"
	weeklyType  = "WEEKLY"
)

func expandPerformanceDiagnostics(ctx context.Context, pd types.Object, diags *diag.Diagnostics) *postgresql.PerformanceDiagnostics {
	if pd.IsNull() || pd.IsUnknown() {
		return nil
	}
	var pdConf PerformanceDiagnostics

	diags.Append(pd.As(ctx, &pdConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &postgresql.PerformanceDiagnostics{
		Enabled:                    pdConf.Enabled.ValueBool(),
		SessionsSamplingInterval:   pdConf.SessionsSamplingInterval.ValueInt64(),
		StatementsSamplingInterval: pdConf.StatementsSamplingInterval.ValueInt64(),
	}
}

func expandBackupRetainPeriodDays(ctx context.Context, cfgBws types.Int64, diags *diag.Diagnostics) *wrapperspb.Int64Value {
	var pgBws *wrapperspb.Int64Value
	if !cfgBws.IsNull() && !cfgBws.IsUnknown() {
		pgBws = &wrapperspb.Int64Value{
			Value: cfgBws.ValueInt64(),
		}
	}

	return pgBws
}

func expandEnvironment(_ context.Context, e types.String, diags *diag.Diagnostics) postgresql.Cluster_Environment {

	if e.IsNull() || e.IsUnknown() {
		return 0
	}

	v, ok := postgresql.Cluster_Environment_value[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(postgresql.Cluster_Environment_value))
		for k, v := range postgresql.Cluster_Environment_value {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		diags.AddError(
			"Failed to parse PostgreSQL environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)

		return 0
	}
	return postgresql.Cluster_Environment(v)
}

func expandBoolWrapper(_ context.Context, b types.Bool, _ *diag.Diagnostics) *wrapperspb.BoolValue {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}

	return wrapperspb.Bool(b.ValueBool())
}

func expandSecurityGroupIds(ctx context.Context, sg types.Set, diags *diag.Diagnostics) []string {
	var securityGroupIds []string
	if !(sg.IsUnknown() || sg.IsNull()) {
		securityGroupIds = make([]string, len(sg.Elements()))
		diags.Append(sg.ElementsAs(ctx, &securityGroupIds, false)...)
		if diags.HasError() {
			return nil
		}
	}

	return securityGroupIds
}

var pgVersionConfigs = map[string]postgresql.ConfigSpec_PostgresqlConfig{
	"10":    &postgresql.ConfigSpec_PostgresqlConfig_10{},
	"10-1c": &postgresql.ConfigSpec_PostgresqlConfig_10_1C{},
	"11":    &postgresql.ConfigSpec_PostgresqlConfig_11{},
	"11-1c": &postgresql.ConfigSpec_PostgresqlConfig_11_1C{},
	"12":    &postgresql.ConfigSpec_PostgresqlConfig_12{},
	"12-1c": &postgresql.ConfigSpec_PostgresqlConfig_12_1C{},
	"13":    &postgresql.ConfigSpec_PostgresqlConfig_13{},
	"13-1c": &postgresql.ConfigSpec_PostgresqlConfig_13_1C{},
	"14":    &postgresql.ConfigSpec_PostgresqlConfig_14{},
	"14-1c": &postgresql.ConfigSpec_PostgresqlConfig_14_1C{},
	"15":    &postgresql.ConfigSpec_PostgresqlConfig_15{},
	"15-1c": &postgresql.ConfigSpec_PostgresqlConfig_15_1C{},
	"16":    &postgresql.ConfigSpec_PostgresqlConfig_16{},
	"16-1c": &postgresql.ConfigSpec_PostgresqlConfig_16_1C{},
	"17":    &postgresql.ConfigSpec_PostgresqlConfig_17{},
	"17-1c": &postgresql.ConfigSpec_PostgresqlConfig_17_1C{},
}

func expandPostgresqlConfig(
	ctx context.Context,
	version string, config mdbcommon.SettingsMapValue,
	diags *diag.Diagnostics,
) postgresql.ConfigSpec_PostgresqlConfig {

	a := protobuf_adapter.NewProtobufMapDataAdapter()

	if pgVersionConfigs[version] == nil {
		diags.AddError("Failed to expand PostgreSQL config.", fmt.Sprintf("unsupported version %s.", version))
		return nil
	}

	pgConf := reflect.New(reflect.TypeOf(pgVersionConfigs[version]).Elem()).Interface()
	if diags.HasError() {
		return nil
	}

	attrs := config.PrimitiveElements(ctx, diags)
	a.Fill(ctx, pgConf, attrs, diags)

	return pgConf.(postgresql.ConfigSpec_PostgresqlConfig)
}

func expandPoolerConfig(ctx context.Context, pCfg types.Object, diags *diag.Diagnostics) *postgresql.ConnectionPoolerConfig {

	pc := &postgresql.ConnectionPoolerConfig{}

	var pcModel PoolerConfig
	diags.Append(pCfg.As(ctx, &pcModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})...)
	if diags.HasError() {
		return nil
	}

	if pd := pcModel.PoolDiscard; !pd.IsNull() && !pd.IsUnknown() {
		pc.SetPoolDiscard(wrapperspb.Bool(pd.ValueBool()))
	}

	if pm := pcModel.PoolingMode; !pm.IsNull() && !pm.IsUnknown() {
		pc.SetPoolingMode(
			postgresql.ConnectionPoolerConfig_PoolingMode(
				postgresql.ConnectionPoolerConfig_PoolingMode_value[pm.ValueString()],
			),
		)
	}

	return pc
}

// TODO: send to api not null structure when fix api
func expandDiskSizeAutoscaling(ctx context.Context, diskSizeAutoscaling types.Object, diags *diag.Diagnostics) *postgresql.DiskSizeAutoscaling {
	if diskSizeAutoscaling.IsNull() || diskSizeAutoscaling.IsUnknown() {
		return nil
	}

	var ds DiskSizeAutoscaling
	if diags.Append(diskSizeAutoscaling.As(ctx, &ds, datasize.DefaultOpts)...); diags.HasError() {
		return nil
	}

	// set attributes PlannedUsageThreshold or EmergencyUsageThreshold to 0 if null
	return &postgresql.DiskSizeAutoscaling{
		DiskSizeLimit:           datasize.ToBytes(ds.DiskSizeLimit.ValueInt64()),
		EmergencyUsageThreshold: ds.EmergencyUsageThreshold.ValueInt64(),
		PlannedUsageThreshold:   ds.PlannedUsageThreshold.ValueInt64(),
	}
}

func expandConfig(ctx context.Context, c types.Object, diags *diag.Diagnostics) *postgresql.ConfigSpec {
	var configSpec Config
	diags.Append(c.As(ctx, &configSpec, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &postgresql.ConfigSpec{
		Version:                configSpec.Version.ValueString(),
		Resources:              mdbcommon.ExpandResources[postgresql.Resources](ctx, configSpec.Resources, diags),
		Autofailover:           expandBoolWrapper(ctx, configSpec.Autofailover, diags),
		Access:                 expandAccess(ctx, configSpec.Access, diags),
		PerformanceDiagnostics: expandPerformanceDiagnostics(ctx, configSpec.PerformanceDiagnostics, diags),
		BackupRetainPeriodDays: expandBackupRetainPeriodDays(ctx, configSpec.BackupRetainPeriodDays, diags),
		BackupWindowStart:      mdbcommon.ExpandBackupWindow(ctx, configSpec.BackupWindowStart, diags),
		PostgresqlConfig:       expandPostgresqlConfig(ctx, configSpec.Version.ValueString(), configSpec.PostgtgreSQLConfig, diags),
		PoolerConfig:           expandPoolerConfig(ctx, configSpec.PoolerConfig, diags),
		DiskSizeAutoscaling:    expandDiskSizeAutoscaling(ctx, configSpec.DiskSizeAutoscaling, diags),
	}
}

func expandFolderId(ctx context.Context, f types.String, providerConfig *config.State, diags *diag.Diagnostics) string {
	folderID, d := validate.FolderID(f, providerConfig)
	diags.Append(d)
	return folderID
}
