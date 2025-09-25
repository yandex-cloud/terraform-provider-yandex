package trino_access_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_access_control/models"
)

var (
	_ datasource.DataSource              = &trinoAccessControlDatasource{}
	_ datasource.DataSourceWithConfigure = &trinoAccessControlDatasource{}
)

func NewDatasource() datasource.DataSource {
	return &trinoAccessControlDatasource{}
}

type trinoAccessControlDatasource struct {
	providerConfig *provider_config.Config
}

// Configure implements datasource.DataSourceWithConfigure.
func (t *trinoAccessControlDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	t.providerConfig = providerConfig
}

// Metadata implements datasource.DataSourceWithConfigure.
func (t *trinoAccessControlDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trino_access_control"
}

// Read implements datasource.DataSourceWithConfigure.
func (t *trinoAccessControlDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state models.AccessControlModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterId.ValueString()
	if clusterID == "" {
		resp.Diagnostics.AddError(
			"Missing cluster_id",
			"cluster_id is required to read Trino access control",
		)
		return
	}

	tflog.Debug(ctx, "Reading Trino access control", clusterIDLogField(clusterID))
	accessControl, diags := GetClusterAccessControl(ctx, t.providerConfig.SDK, clusterID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if accessControl == nil {
		resp.Diagnostics.AddError(
			"Access control not found",
			fmt.Sprintf("Trino cluster %q has no configured access control", clusterID),
		)
		return
	}

	readState, diags := models.FromAPI(ctx, clusterID, accessControl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(state.ApplyChanges(readState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading Trino access control", clusterIDLogField(clusterID))
}

// Schema implements datasource.DataSourceWithConfigure.
func (t *trinoAccessControlDatasource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AccessControlDatasourceSchema(ctx)
}
