package metastore_cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/metastore/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateClusterRequest(ctx context.Context, plan *ClusterModel, providerConfig *config.State) (*metastore.CreateClusterRequest, diag.Diagnostics) {
	diags := &diag.Diagnostics{}

	clusterCreateRequest := &metastore.CreateClusterRequest{
		FolderId:           mdbcommon.ExpandFolderId(ctx, plan.FolderId, providerConfig, diags),
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             mdbcommon.ExpandLabels(ctx, plan.Labels, diags),
		MinServersPerZone:  1,
		MaxServersPerZone:  1,
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		ConfigSpec: &metastore.ConfigSpec{
			Resources: &metastore.Resources{
				ResourcePresetId: plan.ClusterConfig.ResourcePresetId.ValueString(),
			},
		},
		ServiceAccountId: plan.ServiceAccountId.ValueString(),
		Logging:          expandLogging(plan.Logging, diags),
		Network: &metastore.NetworkConfig{
			SubnetIds:        expandCollection[types.Set, []string](ctx, plan.SubnetIds, diags),
			SecurityGroupIds: expandCollection[types.Set, []string](ctx, plan.SecurityGroupIds, diags),
		},
		MaintenanceWindow: expandMaintenanceWindow(plan.MaintenanceWindow, diags),
	}

	return clusterCreateRequest, *diags
}

func BuildUpdateClusterRequest(ctx context.Context, state *ClusterModel, plan *ClusterModel) (*metastore.UpdateClusterRequest, diag.Diagnostics) {
	diags := &diag.Diagnostics{}
	updateMaskPaths := make([]string, 0)

	updateClusterRequest := &metastore.UpdateClusterRequest{
		ClusterId: plan.Id.ValueString(),
	}

	if !stringsAreEqual(plan.Name, state.Name) {
		updateClusterRequest.SetName(plan.Name.ValueString())
		updateMaskPaths = append(updateMaskPaths, "name")
	}

	if !stringsAreEqual(plan.Description, state.Description) {
		updateClusterRequest.SetDescription(plan.Description.ValueString())
		updateMaskPaths = append(updateMaskPaths, "description")
	}

	if !mapsAreEqual(plan.Labels, state.Labels) {
		updateClusterRequest.SetLabels(mdbcommon.ExpandLabels(ctx, plan.Labels, diags))
		updateMaskPaths = append(updateMaskPaths, "labels")
	}

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		updateClusterRequest.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		updateMaskPaths = append(updateMaskPaths, "deletion_protection")
	}

	if !plan.ClusterConfig.Equal(state.ClusterConfig) {
		updateClusterRequest.SetConfigSpec(&metastore.UpdateClusterConfigSpec{
			Resources: &metastore.Resources{
				ResourcePresetId: plan.ClusterConfig.ResourcePresetId.ValueString(),
			},
		})
		updateMaskPaths = append(updateMaskPaths, "config_spec")
	}

	if !plan.ServiceAccountId.Equal(state.ServiceAccountId) {
		updateClusterRequest.SetServiceAccountId(plan.ServiceAccountId.ValueString())
		updateMaskPaths = append(updateMaskPaths, "service_account_id")
	}

	if !loggingValuesAreEqual(plan.Logging, state.Logging) {
		updateClusterRequest.SetLogging(expandLogging(plan.Logging, diags))
		updateMaskPaths = append(updateMaskPaths, "logging")
	}

	if !setsAreEqual(plan.SecurityGroupIds, state.SecurityGroupIds) {
		updateClusterRequest.SetNetworkSpec(&metastore.UpdateNetworkConfigSpec{
			SecurityGroupIds: expandCollection[types.Set, []string](ctx, plan.SecurityGroupIds, diags),
		})
		updateMaskPaths = append(updateMaskPaths, "network_spec.security_group_ids")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		updateClusterRequest.SetMaintenanceWindow(expandMaintenanceWindow(plan.MaintenanceWindow, diags))
		updateMaskPaths = append(updateMaskPaths, "maintenance_window")
	}

	updateClusterRequest.SetUpdateMask(&fieldmaskpb.FieldMask{
		Paths: updateMaskPaths,
	})

	return updateClusterRequest, *diags
}

func logLevelToAPI(minLevelValue types.String) (logging.LogLevel_Level, diag.Diagnostic) {
	if minLevelValue.IsNull() {
		return logging.LogLevel_LEVEL_UNSPECIFIED, nil
	}

	minLevel, ok := logging.LogLevel_Level_value[minLevelValue.ValueString()]
	if !ok || minLevel == 0 {
		return 0, diag.NewErrorDiagnostic("Invalid Metastore cluster logging configuration",
			fmt.Sprintf("Unsupported value for `min_level` attribute provided. It must be one of `%s`", strings.Join(allowedLogLevels(), "`, `")))
	}
	return logging.LogLevel_Level(minLevel), nil
}

func setsAreEqual(set1, set2 types.Set) bool {
	if set1.Equal(set2) {
		return true
	}
	// if one of sets is null and the other is empty then we assume that they are equal
	if len(set1.Elements()) == 0 && len(set2.Elements()) == 0 {
		return true
	}
	return false
}

func mapsAreEqual(map1, map2 types.Map) bool {
	if map1.Equal(map2) {
		return true
	}
	// if one of map is null and the other is empty then we assume that they are equal
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}
	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.Equal(str2) {
		return true
	}
	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	return false
}

func (v LoggingValue) IsExplicitlyDisabled() bool {
	return !v.IsNull() && !v.Enabled.ValueBool()
}

func loggingValuesAreEqual(val1, val2 LoggingValue) bool {
	if val1.Equal(val2) {
		return true
	}
	// if one of values is null and the other is empty then we assume that they are equal
	if (val1.IsExplicitlyDisabled() && val2.IsNull()) || (val1.IsNull() && val2.IsExplicitlyDisabled()) {
		return true
	}

	return false
}

type TfCollection interface {
	attr.Value
	ElementsAs(ctx context.Context, target interface{}, allowUnhandled bool) diag.Diagnostics
}

func expandCollection[Src TfCollection, Dst map[string]string | []string](ctx context.Context, collection Src, diags *diag.Diagnostics) Dst {
	var res Dst
	if utils.IsPresent(collection) {
		diags.Append(collection.ElementsAs(ctx, &res, false)...)
		if diags.HasError() {
			return nil
		}
	}
	return res
}

func expandLogging(logging LoggingValue, diags *diag.Diagnostics) *metastore.LoggingConfig {
	if logging.IsNull() || !logging.Enabled.ValueBool() {
		return nil
	}

	minLevel, d := logLevelToAPI(logging.MinLevel)
	diags.Append(d)
	if diags.HasError() {
		return nil
	}

	loggingConfig := &metastore.LoggingConfig{
		Enabled:  logging.Enabled.ValueBool(),
		MinLevel: minLevel,
	}

	// both folder_id and log_group_id are specified or both are not specified
	if (logging.FolderId.IsNull() || logging.FolderId.IsUnknown()) == (logging.LogGroupId.IsNull() || logging.LogGroupId.IsUnknown()) {
		diags.AddError("Invalid Metastore cluster logging configuration",
			"Exactly one of the attributes `folder_id` and `log_group_id` must be specified")
		return nil
	}

	if !logging.FolderId.IsNull() {
		loggingConfig.Destination = &metastore.LoggingConfig_FolderId{
			FolderId: logging.FolderId.ValueString(),
		}
	} else {
		loggingConfig.Destination = &metastore.LoggingConfig_LogGroupId{
			LogGroupId: logging.LogGroupId.ValueString(),
		}
	}

	return loggingConfig
}

func expandMaintenanceWindow(mw MaintenanceWindowValue, diags *diag.Diagnostics) *metastore.MaintenanceWindow {
	if mw.IsNull() {
		return nil
	}

	maintenanceWindow := &metastore.MaintenanceWindow{}

	switch mw.MaintenanceWindowType.ValueString() {
	case "ANYTIME":
		if !mw.Day.IsNull() || !mw.Hour.IsNull() {
			diags.AddError(
				"Invalid Metastore maintenance window configuration",
				"Any of attributes `day` and `hour` must not be specified for `ANYTIME` window type",
			)
			return nil
		}
		maintenanceWindow.SetAnytime(&metastore.AnytimeMaintenanceWindow{})
	case "WEEKLY":
		if mw.Day.IsNull() || mw.Hour.IsNull() {
			diags.AddError(
				"Invalid Metastore maintenance window configuration",
				"Attributes `day` and `hour` booth must be specified for `WEEKLY` window type",
			)
			return nil
		}

		day := mw.Day.ValueString()
		maintenanceWindow.SetWeeklyMaintenanceWindow(&metastore.WeeklyMaintenanceWindow{
			Day:  metastore.WeeklyMaintenanceWindow_WeekDay(metastore.WeeklyMaintenanceWindow_WeekDay_value[day]),
			Hour: mw.Hour.ValueInt64(),
		})
	default:
		diags.AddError(
			"Invalid Metastore maintenance window configuration",
			fmt.Sprintf("Type must be `ANYTIME` or `WEEKLY`, but '%s' given", mw.MaintenanceWindowType.ValueString()),
		)
		return nil
	}

	return maintenanceWindow
}
