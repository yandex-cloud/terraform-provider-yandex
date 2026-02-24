package mdb_sharded_postgresql_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func expandConfig(ctx context.Context, configSpec Config, diags *diag.Diagnostics) *spqr.ConfigSpec {
	return &spqr.ConfigSpec{
		Access:                 mdbcommon.ExpandAccess[*spqr.Access](ctx, configSpec.Access, diags),
		BackupRetainPeriodDays: expandBackupRetainPeriodDays(ctx, configSpec.BackupRetainPeriodDays, diags),
		BackupWindowStart:      mdbcommon.ExpandBackupWindow(ctx, configSpec.BackupWindowStart, diags),
		SpqrSpec:               expandSPQRConfig(ctx, configSpec.SPQRConfig, diags),
	}
}

func expandBackupRetainPeriodDays(ctx context.Context, cfgBws types.Int64, diags *diag.Diagnostics) *wrapperspb.Int64Value {
	var bws *wrapperspb.Int64Value
	if !cfgBws.IsNull() && !cfgBws.IsUnknown() {
		bws = &wrapperspb.Int64Value{
			Value: cfgBws.ValueInt64(),
		}
	}

	return bws
}

func expandSPQRConfig(
	ctx context.Context,
	config ShardedPostgreSQLConfig,
	diags *diag.Diagnostics,
) *spqr.SpqrSpec {

	a := protobuf_adapter.NewProtobufMapDataAdapter()

	conf := &spqr.SpqrSpec{
		Balancer:   &spqr.BalancerSettings{},
		Postgresql: nil,
	}

	if !config.Common.IsNull() && !config.Common.IsUnknown() {
		attrs := config.Common.PrimitiveElements(ctx, diags)
		if _, ok := attrs["console_password"]; ok {
			diags.AddWarning(`console_password is deprecated`, ConsolePasswordDeprecatedMsg)
		}
		a.FillWithDepth(ctx, conf, attrs, diags, 1)
	}

	if !config.Balancer.IsNull() && !config.Balancer.IsUnknown() {
		attrs := config.Balancer.PrimitiveElements(ctx, diags)
		a.Fill(ctx, conf.Balancer, attrs, diags)
	}

	if config.Router != nil {
		conf.Router = &spqr.SpqrSpec_Router{
			Resources: mdbcommon.ExpandResources[spqr.Resources](ctx, config.Router.Resources, diags),
			Config:    &spqr.RouterSettings{},
		}
		attrs := config.Router.Config.PrimitiveElements(ctx, diags)
		a.Fill(ctx, conf.Router.Config, attrs, diags)
	}

	if config.Coordinator != nil {
		conf.Coordinator = &spqr.SpqrSpec_Coordinator{
			Resources: mdbcommon.ExpandResources[spqr.Resources](ctx, config.Coordinator.Resources, diags),
			Config:    &spqr.CoordinatorSettings{},
		}
		attrs := config.Coordinator.Config.PrimitiveElements(ctx, diags)
		a.Fill(ctx, conf.Coordinator.Config, attrs, diags)
	}

	if config.Infra != nil {
		conf.Infra = &spqr.SpqrSpec_Infra{
			Resources:   mdbcommon.ExpandResources[spqr.Resources](ctx, config.Infra.Resources, diags),
			Router:      &spqr.RouterSettings{},
			Coordinator: &spqr.CoordinatorSettings{},
		}
		attrs := config.Infra.Router.PrimitiveElements(ctx, diags)
		a.Fill(ctx, conf.Infra.Router, attrs, diags)
		attrs = config.Infra.Coordinator.PrimitiveElements(ctx, diags)
		a.Fill(ctx, conf.Infra.Coordinator, attrs, diags)
	}

	return conf
}

func expandLabels(ctx context.Context, labels types.Map, diags *diag.Diagnostics) map[string]string {
	var lMap map[string]string
	if !(labels.IsUnknown() || labels.IsNull()) {
		diags.Append(labels.ElementsAs(ctx, &lMap, false)...)
		if diags.HasError() {
			return nil
		}
	}
	return lMap
}

const (
	anytimeType = "ANYTIME"
	weeklyType  = "WEEKLY"
)

func expandClusterMaintenanceWindow(ctx context.Context, mw types.Object, diags *diag.Diagnostics) *spqr.MaintenanceWindow {
	if mw.IsNull() || mw.IsUnknown() {
		return nil
	}

	out := &spqr.MaintenanceWindow{}
	var mwConf MaintenanceWindow

	diags.Append(mw.As(ctx, &mwConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	if mwType := mwConf.Type.ValueString(); mwType == anytimeType {
		out.Policy = &spqr.MaintenanceWindow_Anytime{
			Anytime: &spqr.AnytimeMaintenanceWindow{},
		}
	} else if mwType == weeklyType {
		mwDay, mwHour := mwConf.Day.ValueString(), mwConf.Hour.ValueInt64()
		day := postgresql.WeeklyMaintenanceWindow_WeekDay_value[mwDay]

		out.Policy = &spqr.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &spqr.WeeklyMaintenanceWindow{
				Hour: mwHour,
				Day:  spqr.WeeklyMaintenanceWindow_WeekDay(day),
			},
		}
	} else {
		diags.AddError(
			"Failed to expand maintenance window.",
			fmt.Sprintf("maintenance_window.type should be %s or %s", anytimeType, weeklyType),
		)
		return nil
	}

	return out
}
