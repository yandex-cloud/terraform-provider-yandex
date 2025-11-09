package cdn_resource

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSource              = &cdnResourceDataSource{}
	_ datasource.DataSourceWithConfigure = &cdnResourceDataSource{}
)

type cdnResourceDataSource struct {
	providerConfig *provider_config.Config
}

// NewDataSource creates a new CDN resource data source
func NewDataSource() datasource.DataSource {
	return &cdnResourceDataSource{}
}

func (d *cdnResourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_resource"
}

func (d *cdnResourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceCDNResourceSchema()
}

func (d *cdnResourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cdnResourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CDNResourceDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading CDN resource data source")

	// Resolve folder_id
	folderID, diag := validate.FolderID(state.FolderID, &d.providerConfig.ProviderState)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine resource ID: either from resource_id or resolve by cname
	var resourceID string
	if !state.ResourceID.IsNull() && state.ResourceID.ValueString() != "" {
		resourceID = state.ResourceID.ValueString()
		tflog.Debug(ctx, "Using provided resource_id", map[string]interface{}{
			"resource_id": resourceID,
		})
	} else if !state.Cname.IsNull() && state.Cname.ValueString() != "" {
		// Resolve by cname using iterator
		resolvedID, err := d.resolveCDNResourceIDByCname(ctx, folderID, state.Cname.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve CDN resource by cname",
				fmt.Sprintf("Error resolving CDN resource by cname %q: %s", state.Cname.ValueString(), err),
			)
			return
		}
		resourceID = resolvedID
		tflog.Debug(ctx, "Resolved resource_id by cname", map[string]interface{}{
			"cname":       state.Cname.ValueString(),
			"resource_id": resourceID,
		})
	} else {
		resp.Diagnostics.AddError(
			"Missing required parameter",
			"Either resource_id or cname must be specified",
		)
		return
	}

	// Fetch CDN resource from API
	tflog.Debug(ctx, "Fetching CDN resource from API", map[string]interface{}{
		"resource_id": resourceID,
	})

	resource, err := d.providerConfig.SDK.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: resourceID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			resp.Diagnostics.AddError(
				"CDN resource not found",
				fmt.Sprintf("CDN resource with ID %s was not found", resourceID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read CDN resource",
			fmt.Sprintf("Error reading CDN resource: %s", err),
		)
		return
	}

	// Convert API response to state
	state.ID = types.StringValue(resource.Id)
	state.ResourceID = types.StringValue(resource.Id)
	state.Cname = types.StringValue(resource.Cname)
	state.FolderID = types.StringValue(resource.FolderId)
	state.Active = types.BoolValue(resource.Active)
	state.CreatedAt = types.StringValue(resource.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z"))
	state.UpdatedAt = types.StringValue(resource.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z"))

	// Set provider type
	if resource.ProviderType != "" {
		state.ProviderType = types.StringValue(resource.ProviderType)
	} else {
		state.ProviderType = types.StringNull()
	}

	// Set provider cname
	if resource.ProviderCname != "" {
		state.ProviderCname = types.StringValue(resource.ProviderCname)
	} else {
		state.ProviderCname = types.StringNull()
	}

	// Set origin protocol
	state.OriginProtocol = flattenOriginProtocol(ctx, resource.OriginProtocol, &resp.Diagnostics)

	// Set origin group ID
	if originGroupID := resource.GetOriginGroupId(); originGroupID != 0 {
		state.OriginGroupID = types.StringValue(strconv.FormatInt(originGroupID, 10))
	} else {
		state.OriginGroupID = types.StringNull()
	}

	// Set origin group name
	state.OriginGroupName = types.StringValue(resource.OriginGroupName)

	// Flatten labels
	if len(resource.Labels) > 0 {
		labels, diags := types.MapValueFrom(ctx, types.StringType, resource.Labels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Labels = labels
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	// Flatten secondary hostnames
	if len(resource.SecondaryHostnames) > 0 {
		secondaryHostnames, diags := types.SetValueFrom(ctx, types.StringType, resource.SecondaryHostnames)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.SecondaryHostnames = secondaryHostnames
	} else {
		state.SecondaryHostnames = types.SetNull(types.StringType)
	}

	// Flatten SSL certificate using existing function
	state.SSLCertificate = flattenSSLCertificate(ctx, resource.SslCertificate, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Shielding - fetch from separate API
	shieldingLocation, err := getShieldingLocation(ctx, resourceID, d.providerConfig.SDK)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read shielding configuration", err.Error())
		return
	}
	state.Shielding = flattenShielding(shieldingLocation)

	// Flatten options using existing function
	// Pass null plan options for data source (no disabled block preservation needed)
	state.Options = FlattenCDNResourceOptions(ctx, resource.Options, types.List{}, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Successfully read CDN resource data source", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// resolveCDNResourceIDByCname resolves CDN resource ID by cname using iterator
func (d *cdnResourceDataSource) resolveCDNResourceIDByCname(ctx context.Context, folderID, cname string) (string, error) {
	if cname == "" {
		return "", fmt.Errorf("empty cname for CDN resource")
	}

	tflog.Debug(ctx, "Resolving CDN resource ID by cname", map[string]interface{}{
		"folder_id": folderID,
		"cname":     cname,
	})

	iterator := d.providerConfig.SDK.CDN().Resource().ResourceIterator(ctx, &cdn.ListResourcesRequest{
		FolderId: folderID,
	})

	for iterator.Next() {
		resource := iterator.Value()
		if cname == resource.Cname {
			tflog.Debug(ctx, "Found matching CDN resource", map[string]interface{}{
				"cname": cname,
				"id":    resource.Id,
			})
			return resource.Id, nil
		}
	}

	if err := iterator.Error(); err != nil {
		return "", fmt.Errorf("error iterating CDN resources: %w", err)
	}

	return "", fmt.Errorf("CDN resource with cname %q not found in folder %s", cname, folderID)
}
