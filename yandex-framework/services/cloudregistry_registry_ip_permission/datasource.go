package cloudregistry_registry_ip_permission

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const yandexCloudRegistryIPPermissionDefaultTimeout = 10 * time.Minute

var _ datasource.DataSourceWithConfigure = (*yandexCloudregistryIPPermissionDataSource)(nil)

type yandexCloudregistryIPPermissionDataSource struct {
	providerConfig *provider_config.Config
}

func NewDataSource() datasource.DataSource {
	return &yandexCloudregistryIPPermissionDataSource{}
}

func (r *yandexCloudregistryIPPermissionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "yandex_cloudregistry_registry_ip_permission"
}

func (r *yandexCloudregistryIPPermissionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	r.providerConfig = providerConfig
}

func (r *yandexCloudregistryIPPermissionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = YandexCloudregistryIPPermissionDataSourceSchema(ctx)
}

func (r *yandexCloudregistryIPPermissionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state yandexCloudregistryIPPermissionDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, timeoutInitError := state.Timeouts.Read(ctx, yandexCloudRegistryIPPermissionDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	id := state.RegistryId.ValueString()
	if state.RegistryId.IsUnknown() || state.RegistryId.IsNull() {
		registryId, d := objectid.ResolveByNameAndFolderID(ctx, r.providerConfig.SDK, r.providerConfig.ProviderState.FolderID.ValueString(), state.RegistryName.ValueString(), sdkresolvers.CloudRegistryResolver)
		resp.Diagnostics.Append(d)
		if resp.Diagnostics.HasError() {
			return
		}
		id = registryId
	}

	reqApi := &cloudregistry.ListIpPermissionsRequest{
		RegistryId: id,
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission request: %s", validate.ProtoDump(reqApi)))

	md := new(metadata.MD)
	res, err := r.providerConfig.SDK.CloudRegistry().Registry().ListIpPermissions(ctx, reqApi, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-request-id: %s", traceHeader[0]))
	}

	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.Diagnostics.AddWarning(
				"Failed to Read resource",
				"IP permission not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed to Read resource",
				"Failed to Read	Error while requesting API to get IP permission: "+err.Error(),
			)
		}
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission response: %s", validate.ProtoDump(res)))

	if res == nil {
		resp.Diagnostics.AddError("Failed to read", "Resource not found")
		return
	}

	newState := flattenYandexCloudregistryIPPermissionDataSource(ctx, res.GetPermissions(), state, state.RegistryName, types.StringValue(id), state.Timeouts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}
