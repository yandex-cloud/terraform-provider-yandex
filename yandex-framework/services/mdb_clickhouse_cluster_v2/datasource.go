package mdb_clickhouse_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

type bindingDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &bindingDataSource{}
}

func (d *bindingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_clickhouse_cluster_v2"
}

func (d *bindingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *bindingDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceClusterSchema(ctx)
}

func (d *bindingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read config into state. Important for default values ​​(e.g. "timeouts").
	var state models.Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := ""
	if !state.ClusterId.IsNull() && !state.ClusterId.IsUnknown() {
		clusterID = state.ClusterId.ValueString()
	}
	name := ""
	if !state.Name.IsNull() && !state.Name.IsUnknown() {
		name = state.Name.ValueString()
	}

	if clusterID == "" && name == "" {
		resp.Diagnostics.AddError(
			"At least one of cluster_id or name is required",
			"The cluster ID or Name must be specified in the configuration",
		)
		return
	}

	// Get cluster id by name and folder.
	if clusterID == "" {
		folderID, diags := validate.FolderID(state.FolderId, &d.providerConfig.ProviderState)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		resolvedID, diags := objectid.ResolveByNameAndFolderID(
			ctx,
			d.providerConfig.SDK,
			folderID,
			name,
			sdkresolvers.ClickhouseClusterResolver,
		)
		resp.Diagnostics.Append(diags)
		if resp.Diagnostics.HasError() {
			return
		}

		clusterID = resolvedID
		state.ClusterId = types.StringValue(clusterID)
	}

	state.Id = types.StringValue(clusterID)
	prevState := state

	refreshState(ctx, &prevState, &state, d.providerConfig.SDK, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
