package trino_cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func BuildCreateClusterRequest(ctx context.Context, clusterModel *ClusterModel, providerConfig *config.State) (*trino.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(clusterModel.FolderId, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	subnetIds := make([]string, 0, len(clusterModel.SubnetIds.Elements()))
	diags.Append(clusterModel.SubnetIds.ElementsAs(ctx, &subnetIds, false)...)
	if diags.HasError() {
		return nil, diags
	}

	common, _, dd := buildCommonForCreateAndUpdate(ctx, clusterModel, nil)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	clusterCreateRequest := &trino.CreateClusterRequest{
		FolderId:    folderID,
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		Trino: &trino.TrinoConfigSpec{
			CoordinatorConfig:  common.Coordinator,
			WorkerConfig:       common.Worker,
			RetryPolicy:        common.RetryPolicy,
			Version:            common.Version,
			Tls:                common.Tls,
			ResourceManagement: common.ResourceManagement,
		},
		Network: &trino.NetworkConfig{
			SubnetIds:        subnetIds,
			SecurityGroupIds: common.SecurityGroupIds,
			PrivateAccess:    common.PrivateAccess,
		},
		DeletionProtection: common.DeletionProtection,
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
		MaintenanceWindow:  common.MaintenanceWindow,
	}

	return clusterCreateRequest, diags
}

type CommonForCreateAndUpdate struct {
	Name               string
	Description        string
	Labels             map[string]string
	SecurityGroupIds   []string
	PrivateAccess      *trino.PrivateAccessConfig
	DeletionProtection bool
	ServiceAccountId   string
	Logging            *trino.LoggingConfig

	Coordinator *trino.CoordinatorConfig
	Worker      *trino.WorkerConfig
	Tls         *trino.TLSConfig
	RetryPolicy *trino.RetryPolicyConfig
	Version     string

	MaintenanceWindow  *trino.MaintenanceWindow
	ResourceManagement *trino.ResourceManagementConfig
}

func (c *CommonForCreateAndUpdate) workerConfigForUpdate() *trino.UpdateWorkerConfig {
	scalePolicy := &trino.UpdateWorkerConfig_WorkerScalePolicy{}
	switch scale := c.Worker.ScalePolicy.ScaleType.(type) {
	case *trino.WorkerConfig_WorkerScalePolicy_FixedScale:
		scalePolicy.ScaleType = &trino.UpdateWorkerConfig_WorkerScalePolicy_FixedScale{
			FixedScale: scale.FixedScale,
		}
	case *trino.WorkerConfig_WorkerScalePolicy_AutoScale:
		scalePolicy.ScaleType = &trino.UpdateWorkerConfig_WorkerScalePolicy_AutoScale{
			AutoScale: scale.AutoScale,
		}
	default:
		return nil
	}

	return &trino.UpdateWorkerConfig{
		Resources:   c.Worker.Resources,
		ScalePolicy: scalePolicy,
	}
}

func (c *CommonForCreateAndUpdate) coordinatorConfigForUpdate() *trino.UpdateCoordinatorConfig {
	return &trino.UpdateCoordinatorConfig{
		Resources: c.Coordinator.Resources,
	}
}

func buildCommonForCreateAndUpdate(ctx context.Context, plan, state *ClusterModel) (*CommonForCreateAndUpdate, []string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	updateMaskPaths := make([]string, 0)

	if state != nil {
		if !plan.Name.Equal(state.Name) {
			updateMaskPaths = append(updateMaskPaths, "name")
		}
		if !stringsAreEqual(plan.Description, state.Description) {
			updateMaskPaths = append(updateMaskPaths, "description")
		}
		if !plan.DeletionProtection.Equal(state.DeletionProtection) {
			updateMaskPaths = append(updateMaskPaths, "deletion_protection")
		}
		if !plan.ServiceAccountId.Equal(state.ServiceAccountId) {
			updateMaskPaths = append(updateMaskPaths, "service_account_id")
		}
	}

	labels := make(map[string]string, len(plan.Labels.Elements()))
	diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !mapsAreEqual(plan.Labels, state.Labels) {
		updateMaskPaths = append(updateMaskPaths, "labels")
	}

	securityGroupIds := make([]string, 0, len(plan.SecurityGroupIds.Elements()))
	diags.Append(plan.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if state != nil && !setsAreEqual(plan.SecurityGroupIds, state.SecurityGroupIds) {
		updateMaskPaths = append(updateMaskPaths, "network_spec.security_group_ids")
	}

	privateAccess := &trino.PrivateAccessConfig{
		Enabled: plan.PrivateAccess.ValueBool(),
	}
	if state != nil && !plan.PrivateAccess.Equal(state.PrivateAccess) {
		updateMaskPaths = append(updateMaskPaths, "network_spec.private_access")
	}

	var loggingConfig *trino.LoggingConfig
	if !plan.Logging.IsNull() {
		minLevel, d := logLevelToAPI(plan.Logging.MinLevel)
		diags.Append(d)
		if diags.HasError() {
			return nil, nil, diags
		}

		loggingConfig = &trino.LoggingConfig{
			Enabled:  plan.Logging.Enabled.ValueBool(),
			MinLevel: minLevel,
		}

		// both folder_id and log_group_id are specified or both are not specified
		if plan.Logging.FolderId.IsNull() == plan.Logging.LogGroupId.IsNull() {
			diags.AddError("Invalid Trino cluster logging configuration",
				"Exactly one of the attributes `folder_id` and `log_group_id` must be specified")
			return nil, nil, diags
		}

		if !plan.Logging.FolderId.IsNull() {
			loggingConfig.Destination = &trino.LoggingConfig_FolderId{
				FolderId: plan.Logging.FolderId.ValueString(),
			}
		} else {
			loggingConfig.Destination = &trino.LoggingConfig_LogGroupId{
				LogGroupId: plan.Logging.LogGroupId.ValueString(),
			}
		}
	}
	if state != nil && !loggingValuesAreEqual(plan.Logging, state.Logging) {
		updateMaskPaths = append(updateMaskPaths, "logging")
	}

	coordinatorConfig := &trino.CoordinatorConfig{
		Resources: &trino.Resources{ResourcePresetId: plan.Coordinator.ResourcePresetId.ValueString()},
	}
	if state != nil && !plan.Coordinator.Equal(state.Coordinator) {
		updateMaskPaths = append(updateMaskPaths, "trino.coordinator_config")
	}

	workerConfig := &trino.WorkerConfig{
		Resources: &trino.Resources{ResourcePresetId: plan.Worker.ResourcePresetId.ValueString()},
	}

	if !isNullOrUnknown(plan.Worker.FixedScale) {
		fixedScaleObject, dd := FixedScaleType{}.ValueFromObject(ctx, plan.Worker.FixedScale)
		diags.Append(dd...)
		if diags.HasError() {
			return nil, nil, diags
		}
		fixedScale := fixedScaleObject.(FixedScaleValue)
		if !fixedScale.IsNull() {
			workerConfig.ScalePolicy = &trino.WorkerConfig_WorkerScalePolicy{
				ScaleType: &trino.WorkerConfig_WorkerScalePolicy_FixedScale{
					FixedScale: &trino.FixedScalePolicy{Count: fixedScale.Count.ValueInt64()},
				},
			}
		}
	}

	if !isNullOrUnknown(plan.Worker.AutoScale) {
		autoScaleObject, dd := AutoScaleType{}.ValueFromObject(ctx, plan.Worker.AutoScale)
		diags.Append(dd...)
		if diags.HasError() {
			return nil, nil, diags
		}
		autoScale := autoScaleObject.(AutoScaleValue)
		if !autoScale.IsNull() {
			workerConfig.ScalePolicy = &trino.WorkerConfig_WorkerScalePolicy{
				ScaleType: &trino.WorkerConfig_WorkerScalePolicy_AutoScale{
					AutoScale: &trino.AutoScalePolicy{
						MinCount: autoScale.MinCount.ValueInt64(),
						MaxCount: autoScale.MaxCount.ValueInt64(),
					},
				},
			}
		}
	}
	if state != nil && !plan.Worker.Equal(state.Worker) {
		updateMaskPaths = append(updateMaskPaths, "trino.worker_config")
	}

	var tlsConfig *trino.TLSConfig
	if !isNullOrUnknown(plan.Tls) {
		trustedCertificates := make([]string, len(plan.Tls.TrustedCertificates.Elements()))
		diags.Append(plan.Tls.TrustedCertificates.ElementsAs(ctx, &trustedCertificates, false)...)
		tlsConfig = &trino.TLSConfig{
			TrustedCertificates: trustedCertificates,
		}
	}
	if state != nil && !tlsValuesAreEqual(state.Tls, plan.Tls) {
		updateMaskPaths = append(updateMaskPaths, "trino.tls")
	}

	var retrPolicyConfig *trino.RetryPolicyConfig
	if !isNullOrUnknown(plan.RetryPolicy.ExchangeManager) {
		// ExchangeManager
		ExchangeManagerObject, dd := ExchangeManagerType{}.ValueFromObject(ctx, plan.RetryPolicy.ExchangeManager)
		diags.Append(dd...)
		if diags.HasError() {
			return nil, nil, diags
		}
		exchangeManager := ExchangeManagerObject.(ExchangeManagerValue)
		exchangeManagerAdditionalProperties := make(map[string]string, len(exchangeManager.AdditionalProperties.Elements()))
		diags.Append(exchangeManager.AdditionalProperties.ElementsAs(ctx, &exchangeManagerAdditionalProperties, false)...)

		// RetryPolicy
		additionalProperties := make(map[string]string, len(plan.RetryPolicy.AdditionalProperties.Elements()))
		diags.Append(plan.RetryPolicy.AdditionalProperties.ElementsAs(ctx, &additionalProperties, false)...)

		retrPolicyConfig = &trino.RetryPolicyConfig{
			Policy: trino.RetryPolicyConfig_RetryPolicy(trino.RetryPolicyConfig_RetryPolicy_value[plan.RetryPolicy.Policy.ValueString()]),
			ExchangeManager: &trino.ExchangeManagerConfig{
				AdditionalProperties: exchangeManagerAdditionalProperties,
				Storage: &trino.ExchangeManagerStorage{
					Type: &trino.ExchangeManagerStorage_ServiceS3_{
						ServiceS3: &trino.ExchangeManagerStorage_ServiceS3{},
					},
				},
			},
			AdditionalProperties: additionalProperties,
		}
	}
	if state != nil && !plan.RetryPolicy.Equal(state.RetryPolicy) {
		updateMaskPaths = append(updateMaskPaths, "trino.retry_policy")
	}

	var version string
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		version = plan.Version.ValueString()
	}
	if state != nil && !plan.Version.Equal(state.Version) {
		updateMaskPaths = append(updateMaskPaths, "trino.version")
	}

	var maintenanceWindow *trino.MaintenanceWindow
	if !isNullOrUnknown(plan.MaintenanceWindow.MaintenanceWindowType) {
		maintenanceWindow = &trino.MaintenanceWindow{}

		switch plan.MaintenanceWindow.MaintenanceWindowType.ValueString() {
		case "ANYTIME":
			if !plan.MaintenanceWindow.Day.IsNull() || !plan.MaintenanceWindow.Hour.IsNull() {
				diags.AddError(
					"Invalid Trino maintenance window configuration",
					"Any of attributes `day` and `hour` must not be specified for `ANYTIME` window type",
				)
				return nil, nil, diags
			}
			maintenanceWindow.SetAnytime(&trino.AnytimeMaintenanceWindow{})
		case "WEEKLY":
			if plan.MaintenanceWindow.Day.IsNull() || plan.MaintenanceWindow.Hour.IsNull() {
				diags.AddError(
					"Invalid Trino maintenance window configuration",
					"Attributes `day` and `hour` booth must be specified for `WEEKLY` window type",
				)
				return nil, nil, diags
			}

			day := plan.MaintenanceWindow.Day.ValueString()
			maintenanceWindow.SetWeeklyMaintenanceWindow(&trino.WeeklyMaintenanceWindow{
				Day:  trino.WeeklyMaintenanceWindow_WeekDay(trino.WeeklyMaintenanceWindow_WeekDay_value[day]),
				Hour: plan.MaintenanceWindow.Hour.ValueInt64(),
			})
		default:
			diags.AddError(
				"Invalid Trino maintenance window configuration",
				fmt.Sprintf("Type must be `ANYTIME` or `WEEKLY`, but '%s' given", plan.MaintenanceWindow.MaintenanceWindowType.ValueString()),
			)
			return nil, nil, diags
		}
	}
	if state != nil && !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		updateMaskPaths = append(updateMaskPaths, "maintenance_window")
	}

	// Resource Management configuration
	resourceManagementConfig := &trino.ResourceManagementConfig{}

	newResourceGroupsModel := ResourceGroups{}
	if plan.ResourceGroupsJson.ValueString() != "" {
		if err := json.Unmarshal([]byte(plan.ResourceGroupsJson.ValueString()), &newResourceGroupsModel); err != nil {
			diags.AddError("Failed to unmarshal Resource Groups", err.Error())
			return nil, nil, diags
		}
	}
	if !isNullOrUnknown(plan.ResourceGroupsJson) {
		resourceGroups := newResourceGroupsModel.ToAPI()
		resourceManagementConfig.ResourceGroups = resourceGroups
	}
	if state != nil {
		equal, dd := resourceGroupsAreEqual(state.ResourceGroupsJson, &newResourceGroupsModel)
		diags.Append(dd...)
		if diags.HasError() {
			return nil, nil, diags
		}
		if !equal {
			updateMaskPaths = append(updateMaskPaths, "trino.resource_management.resource_groups")
		}
	}

	queryProperties := make(map[string]string, len(plan.QueryProperties.Elements()))
	diags.Append(plan.QueryProperties.ElementsAs(ctx, &queryProperties, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	if !isNullOrUnknown(plan.QueryProperties) {
		resourceManagementConfig.Query = &trino.QueryConfig{
			Properties: queryProperties,
		}
	}
	if state != nil && !mapsAreEqual(plan.QueryProperties, state.QueryProperties) {
		updateMaskPaths = append(updateMaskPaths, "trino.resource_management.query")
	}

	params := &CommonForCreateAndUpdate{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             labels,
		SecurityGroupIds:   securityGroupIds,
		PrivateAccess:      privateAccess,
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		ServiceAccountId:   plan.ServiceAccountId.ValueString(),
		Logging:            loggingConfig,
		Coordinator:        coordinatorConfig,
		Worker:             workerConfig,
		MaintenanceWindow:  maintenanceWindow,
		Tls:                tlsConfig,
		RetryPolicy:        retrPolicyConfig,
		Version:            version,
		ResourceManagement: resourceManagementConfig,
	}

	return params, updateMaskPaths, diags
}

func logLevelToAPI(minLevelValue types.String) (logging.LogLevel_Level, diag.Diagnostic) {
	if minLevelValue.IsNull() {
		return logging.LogLevel_LEVEL_UNSPECIFIED, nil
	}

	minLevel, ok := logging.LogLevel_Level_value[minLevelValue.ValueString()]
	if !ok || minLevel == 0 {
		return 0, diag.NewErrorDiagnostic("Invalid Trino cluster logging configuration",
			fmt.Sprintf("Unsupported value for `min_level` attribute provided. It must be one of `%s`", strings.Join(allowedLogLevels(), "`, `")))
	}
	return logging.LogLevel_Level(minLevel), nil
}

func BuildUpdateClusterRequest(ctx context.Context, state *ClusterModel, plan *ClusterModel) (*trino.UpdateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	common, updateMaskPaths, dd := buildCommonForCreateAndUpdate(ctx, plan, state)
	diags.Append(dd...)
	if diags.HasError() {
		return nil, diags
	}

	updateClusterRequest := &trino.UpdateClusterRequest{
		ClusterId:   state.Id.ValueString(),
		UpdateMask:  &field_mask.FieldMask{Paths: updateMaskPaths},
		Name:        common.Name,
		Description: common.Description,
		Labels:      common.Labels,
		Trino: &trino.UpdateTrinoConfigSpec{
			CoordinatorConfig:  common.coordinatorConfigForUpdate(),
			WorkerConfig:       common.workerConfigForUpdate(),
			RetryPolicy:        common.RetryPolicy,
			Version:            common.Version,
			Tls:                common.Tls,
			ResourceManagement: common.ResourceManagement,
		},
		NetworkSpec: &trino.UpdateNetworkConfigSpec{
			SecurityGroupIds: common.SecurityGroupIds,
			PrivateAccess:    common.PrivateAccess,
		},
		DeletionProtection: common.DeletionProtection,
		ServiceAccountId:   common.ServiceAccountId,
		Logging:            common.Logging,
		MaintenanceWindow:  common.MaintenanceWindow,
	}

	return updateClusterRequest, diags
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

func tlsValuesAreEqual(a, b TlsValue) bool {
	if a.Equal(b) {
		return true
	}
	return isEmptyTlsValue(a) && isEmptyTlsValue(b)
}

func isEmptyTlsValue(t TlsValue) bool {
	return t.IsNull() || (!t.IsUnknown() && len(t.TrustedCertificates.Elements()) == 0)
}
