package airflow_cluster

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ datasource.DataSource              = &airflowClusterDatasource{}
	_ datasource.DataSourceWithConfigure = &airflowClusterDatasource{}
)

func NewDatasource() datasource.DataSource {
	return &airflowClusterDatasource{}
}

type airflowClusterDatasource struct {
	providerConfig *provider_config.Config
}

func (a *airflowClusterDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_airflow_cluster"
}

func (a *airflowClusterDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ClusterDataSourceSchema(ctx)
}

func (a *airflowClusterDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ClusterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	if id == "" {
		folderID, d := validate.FolderID(state.FolderId, &a.providerConfig.ProviderState)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		id, d = objectid.ResolveByNameAndFolderID(ctx, a.providerConfig.SDK, folderID, state.Name.ValueString(), sdkresolvers.AirflowClusterResolver)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Id = types.StringValue(id)
	}

	updateState(ctx, a.providerConfig.SDK, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (a *airflowClusterDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.providerConfig = providerConfig
}
