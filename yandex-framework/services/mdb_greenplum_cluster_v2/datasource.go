package mdb_greenplum_cluster_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	greenplumv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var _ datasource.DataSourceWithConfigure = (*yandexMdbGreenplumClusterV2DataSource)(nil)

type yandexMdbGreenplumClusterV2DataSource struct {
	providerConfig *providerconfig.Config
}

func NewDataSource() datasource.DataSource {
	return &yandexMdbGreenplumClusterV2DataSource{}
}

func (r *yandexMdbGreenplumClusterV2DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "yandex_mdb_greenplum_cluster_v2"
}

func (r *yandexMdbGreenplumClusterV2DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*providerconfig.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *yandexMdbGreenplumClusterV2DataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = YandexMdbGreenplumClusterV2DatasourceSchema(ctx)
}

func (r *yandexMdbGreenplumClusterV2DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state yandexMdbGreenplumClusterV2DatasourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, timeoutInitError := state.Timeouts.Read(ctx, providerconfig.DefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	reqApi := &greenplum.GetClusterRequest{}
	id := state.ID.ValueString()
	if !state.ID.IsUnknown() && !state.ID.IsNull() {
		id = state.ID.ValueString()
	}
	reqApi.SetClusterId(id)

	tflog.Debug(ctx, fmt.Sprintf("Read cluster request: %s", validate.ProtoDump(reqApi)))

	md := new(metadata.MD)
	res, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Get(ctx, reqApi, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read cluster x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read cluster x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.Diagnostics.AddWarning(
				"Failed to Read resource",
				"cluster not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed to Read resource",
				"Error while requesting API to get cluster:"+err.Error(),
			)
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Read cluster response: %s", validate.ProtoDump(res)))

	if resp.Diagnostics.HasError() {
		return
	}

	// diagnostics don't have errors and resource is nil => resource not found
	if res == nil {
		resp.Diagnostics.AddWarning("Failed to read", "Resource not found")
		return
	}

	newState := flattenYandexMdbGreenplumClusterV2Datasource(ctx, res, state, state.Timeouts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
