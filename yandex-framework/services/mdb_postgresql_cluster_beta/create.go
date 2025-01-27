package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func prepareCreateRequest(ctx context.Context, plan *Cluster, providerConfig *config.State) (*postgresql.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	folderID, d := validate.FolderID(plan.FolderId, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	var labels map[string]string
	if !(plan.Labels.IsUnknown() || plan.Labels.IsNull()) {
		diags.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	var env postgresql.Cluster_Environment
	if !(plan.Environment.IsUnknown() || plan.Environment.IsNull()) {
		env, d = toEnvironment(plan.Environment)
		diags.Append(d)
		if diags.HasError() {
			return nil, diags
		}
	}

	var configSpec Config
	diags.Append(plan.Config.As(ctx, &configSpec, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	var resources Resources
	diags.Append(configSpec.Resources.As(ctx, &resources, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	var securityGroupIds []string
	if !(plan.SecurityGroupIds.IsUnknown() || plan.SecurityGroupIds.IsNull()) {
		securityGroupIds = make([]string, len(plan.SecurityGroupIds.Elements()))
		diags.Append(plan.SecurityGroupIds.ElementsAs(ctx, &securityGroupIds, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	request := &postgresql.CreateClusterRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		FolderId:    folderID,
		NetworkId:   plan.NetworkId.ValueString(),
		Environment: env,
		Labels:      labels,
		ConfigSpec: &postgresql.ConfigSpec{
			Version: configSpec.Version.ValueString(),
			Resources: &postgresql.Resources{
				ResourcePresetId: resources.ResourcePresetID.ValueString(),
				DiskTypeId:       resources.DiskTypeID.ValueString(),
				DiskSize:         datasize.ToBytes(resources.DiskSize.ValueInt64()),
			},
			Autofailover: &wrappers.BoolValue{
				Value: configSpec.Autofailover.ValueBool(),
			},
			Access:                 expandAccess(ctx, configSpec.Access, &diags),
			PerformanceDiagnostics: expandPerformanceDiagnostics(ctx, configSpec.PerformanceDiagnostics, &diags),
			BackupRetainPeriodDays: expandBackupRetainPeriodDays(ctx, configSpec.BackupRetainPeriodDays, &diags),
			BackupWindowStart:      expandBackupWindowStart(ctx, configSpec.BackupWindowStart, &diags),
		},
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		SecurityGroupIds:   securityGroupIds,
	}

	return request, diags
}

func toEnvironment(e basetypes.StringValue) (postgresql.Cluster_Environment, diag.Diagnostic) {
	v, ok := postgresql.Cluster_Environment_value[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(postgresql.Cluster_Environment_value))
		for k, v := range postgresql.Cluster_Environment_value {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		return 0, diag.NewErrorDiagnostic(
			"Failed to parse PostgreSQL environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)
	}
	return postgresql.Cluster_Environment(v), nil
}
