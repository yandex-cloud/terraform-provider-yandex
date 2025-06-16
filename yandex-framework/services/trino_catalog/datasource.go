package trino_catalog

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ datasource.DataSource              = &trinoCatalogDatasource{}
	_ datasource.DataSourceWithConfigure = &trinoCatalogDatasource{}
)

func NewDatasource() datasource.DataSource {
	return &trinoCatalogDatasource{}
}

type trinoCatalogDatasource struct {
	providerConfig *provider_config.Config
}

func (d *trinoCatalogDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trino_catalog"
}

func (d *trinoCatalogDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CatalogDataSourceSchema(ctx)
}

func (d *trinoCatalogDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CatalogModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.ClusterId.ValueString()
	if clusterId == "" {
		resp.Diagnostics.AddError(
			"Missing cluster_id",
			"cluster_id is required to read Trino catalog",
		)
		return
	}

	catalogId := state.Id.ValueString()
	if catalogId == "" {
		// If catalog ID is not provided, try to resolve by name
		catalogName := state.Name.ValueString()
		if catalogName == "" {
			resp.Diagnostics.AddError(
				"Missing catalog identifier",
				"Either id or name must be provided to read Trino catalog",
			)
			return
		}

		var diag diag.Diagnostic
		catalogId, diag = GetCatalogByName(ctx, d.providerConfig.SDK, clusterId, catalogName)
		resp.Diagnostics.Append(diag)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Id = types.StringValue(catalogId)
	}

	catalog, diag := GetCatalogByID(ctx, d.providerConfig.SDK, catalogId, clusterId)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	if catalog == nil {
		resp.Diagnostics.AddError(
			"Catalog not found",
			fmt.Sprintf("Trino catalog with ID %s not found in cluster %s", catalogId, clusterId),
		)
		return
	}

	resp.Diagnostics.Append(updateState(ctx, d.providerConfig.SDK, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *trinoCatalogDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
