package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/request/nodegroups"
)

func PrepareCreateRequest(ctx context.Context, plan *model.OpenSearch, providerConfig *config.State) (*opensearch.CreateClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// o.providerConfig.ProviderState.FolderID -- as default FolderID if not specified
	folderID, d := validate.FolderID(plan.FolderID, providerConfig)
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

	var env opensearch.Cluster_Environment
	if !(plan.Environment.IsUnknown() || plan.Environment.IsNull()) {
		env, d = toEnvironment(plan.Environment)
		diags.Append(d)
		if diags.HasError() {
			return nil, diags
		}
	}

	config, diags := prepareConfigCreateSpec(ctx, plan)
	if diags.HasError() {
		return nil, diags
	}

	// o.providerConfig.ProviderState.Endpoint -- to restrict network_id (in compute network_id is required)
	networkID, d := validate.NetworkId(plan.NetworkID, providerConfig)
	diags.Append(d)
	if diags.HasError() {
		return nil, diags
	}

	var securityGroupIds []string
	if !(plan.SecurityGroupIDs.IsUnknown() || plan.SecurityGroupIDs.IsNull()) {
		diags.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIds, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	mw, diags := prepareMaintenanceWindow(ctx, plan)
	if diags.HasError() {
		return nil, diags
	}

	req := &opensearch.CreateClusterRequest{
		FolderId:           folderID,
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Labels:             labels,
		Environment:        env,
		ConfigSpec:         config,
		NetworkId:          networkID,
		SecurityGroupIds:   securityGroupIds,
		ServiceAccountId:   plan.ServiceAccountID.ValueString(),
		DeletionProtection: plan.DeletionProtection.ValueBool(),
		MaintenanceWindow:  mw,
	}

	return req, diag.Diagnostics{}
}

func toEnvironment(e basetypes.StringValue) (opensearch.Cluster_Environment, diag.Diagnostic) {
	v, ok := opensearch.Cluster_Environment_value[e.ValueString()]
	if !ok || v == 0 {
		allowedEnvs := make([]string, 0, len(opensearch.Cluster_Environment_value))
		for k, v := range opensearch.Cluster_Environment_value {
			if v == 0 {
				continue
			}
			allowedEnvs = append(allowedEnvs, k)
		}

		return 0, diag.NewErrorDiagnostic(
			"Failed to parse OpenSearch environment",
			fmt.Sprintf("Error while parsing value for 'environment'. Value must be one of `%s`, not `%s`", strings.Join(allowedEnvs, "`, `"), e),
		)
	}
	return opensearch.Cluster_Environment(v), nil
}

func prepareConfigCreateSpec(ctx context.Context, c *model.OpenSearch) (*opensearch.ConfigCreateSpec, diag.Diagnostics) {
	config, diags := model.ParseConfig(ctx, c)
	if diags.HasError() {
		return nil, diags
	}

	access, diags := tryToAccess(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	if config.OpenSearch.IsNull() || config.OpenSearch.IsUnknown() {
		diags.AddError("config.opensearch is required", "")
		return nil, diags
	}

	openSearchBlock, diags := model.ParseOpenSearchSubConfig(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	var plugins []string
	if !(openSearchBlock.Plugins.IsUnknown() || openSearchBlock.Plugins.IsNull()) {
		plugins = make([]string, 0, len(openSearchBlock.Plugins.Elements()))
		diags.Append(openSearchBlock.Plugins.ElementsAs(ctx, &plugins, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	nodeGroups, diags := nodegroups.PrepareOpenSearchCreate(ctx, openSearchBlock)
	if diags.HasError() {
		return nil, diags
	}

	opensearchSpec := &opensearch.OpenSearchCreateSpec{
		NodeGroups: nodeGroups,
		Plugins:    plugins,
	}

	if config.Dashboards.IsNull() || config.Dashboards.IsUnknown() {
		return &opensearch.ConfigCreateSpec{
			Access:         access,
			AdminPassword:  config.AdminPassword.ValueString(),
			Version:        config.Version.ValueString(),
			OpensearchSpec: opensearchSpec,
		}, diags
	}

	dashboardsBlock, diags := model.ParseDashboardSubConfig(ctx, config)
	if diags.HasError() {
		return nil, diags
	}

	dashboardsNodeGroups, diags := nodegroups.PrepareDashboardsCreate(ctx, dashboardsBlock)
	if diags.HasError() {
		return nil, diags
	}

	dashboardsSpec := &opensearch.DashboardsCreateSpec{
		NodeGroups: dashboardsNodeGroups,
	}

	return &opensearch.ConfigCreateSpec{
		Access:         access,
		AdminPassword:  config.AdminPassword.ValueString(),
		Version:        config.Version.ValueString(),
		OpensearchSpec: opensearchSpec,
		DashboardsSpec: dashboardsSpec,
	}, diags
}

func tryToAccess(ctx context.Context, cfg *model.Config) (*opensearch.Access, diag.Diagnostics) {
	if cfg.Access.IsUnknown() || cfg.Access.IsNull() {
		return nil, diag.Diagnostics{}
	}

	access, diags := model.ParseAccess(ctx, cfg)
	if diags.HasError() {
		return nil, diags
	}

	return &opensearch.Access{
		DataTransfer: access.DataTransfer.ValueBool(),
		Serverless:   access.Serverless.ValueBool(),
	}, diag.Diagnostics{}
}
