package cdn_origin_group

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
	_ datasource.DataSource              = &cdnOriginGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &cdnOriginGroupDataSource{}
)

type cdnOriginGroupDataSource struct {
	providerConfig *provider_config.Config
}

// NewDataSource creates a new CDN origin group data source
func NewDataSource() datasource.DataSource {
	return &cdnOriginGroupDataSource{}
}

func (d *cdnOriginGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_origin_group"
}

func (d *cdnOriginGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceCDNOriginGroupSchema()
}

func (d *cdnOriginGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cdnOriginGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CDNOriginGroupDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading CDN origin group data source")

	// Resolve folder_id
	folderID, diag := validate.FolderID(state.FolderID, &d.providerConfig.ProviderState)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine origin group ID: either from origin_group_id or resolve by name
	var originGroupID int64
	if !state.OriginGroupID.IsNull() && state.OriginGroupID.ValueString() != "" {
		var err error
		originGroupID, err = strconv.ParseInt(state.OriginGroupID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid origin_group_id format",
				fmt.Sprintf("Error parsing origin_group_id %q: %s", state.OriginGroupID.ValueString(), err),
			)
			return
		}
		tflog.Debug(ctx, "Using provided origin_group_id", map[string]interface{}{
			"origin_group_id": originGroupID,
		})
	} else if !state.Name.IsNull() && state.Name.ValueString() != "" {
		// Resolve by name using iterator
		resolvedID, err := d.resolveCDNOriginGroupIDByName(ctx, folderID, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve origin group by name",
				fmt.Sprintf("Error resolving CDN origin group by name %q: %s", state.Name.ValueString(), err),
			)
			return
		}
		originGroupID = resolvedID
		tflog.Debug(ctx, "Resolved origin_group_id by name", map[string]interface{}{
			"name":            state.Name.ValueString(),
			"origin_group_id": originGroupID,
		})
	} else {
		resp.Diagnostics.AddError(
			"Missing required parameter",
			"Either origin_group_id or name must be specified",
		)
		return
	}

	// Fetch origin group from API
	tflog.Debug(ctx, "Fetching origin group from API", map[string]interface{}{
		"folder_id":       folderID,
		"origin_group_id": originGroupID,
	})

	originGroup, err := d.providerConfig.SDK.CDN().OriginGroup().Get(ctx, &cdn.GetOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: originGroupID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			resp.Diagnostics.AddError(
				"Origin group not found",
				fmt.Sprintf("CDN origin group with ID %d was not found in folder %s", originGroupID, folderID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read origin group",
			fmt.Sprintf("Error reading CDN origin group: %s", err),
		)
		return
	}

	// Convert API response to state
	state.ID = types.StringValue(strconv.FormatInt(originGroup.Id, 10))
	state.OriginGroupID = types.StringValue(strconv.FormatInt(originGroup.Id, 10))
	state.FolderID = types.StringValue(originGroup.FolderId)
	state.Name = types.StringValue(originGroup.Name)
	state.UseNext = types.BoolValue(originGroup.UseNext)

	// Set provider type if available
	if originGroup.ProviderType != "" {
		state.ProviderType = types.StringValue(originGroup.ProviderType)
	} else {
		state.ProviderType = types.StringNull()
	}

	// Flatten origins
	parentGroupID := strconv.FormatInt(originGroup.Id, 10)
	state.Origins = flattenOrigins(ctx, originGroup.Origins, parentGroupID, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Successfully read CDN origin group data source", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// resolveCDNOriginGroupIDByName resolves origin group ID by name using iterator
func (d *cdnOriginGroupDataSource) resolveCDNOriginGroupIDByName(ctx context.Context, folderID, name string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("empty name for origin group")
	}

	tflog.Debug(ctx, "Resolving origin group ID by name", map[string]interface{}{
		"folder_id": folderID,
		"name":      name,
	})

	iterator := d.providerConfig.SDK.CDN().OriginGroup().OriginGroupIterator(ctx, &cdn.ListOriginGroupsRequest{
		FolderId: folderID,
	})

	for iterator.Next() {
		originGroup := iterator.Value()
		if name == originGroup.Name {
			tflog.Debug(ctx, "Found matching origin group", map[string]interface{}{
				"name": name,
				"id":   originGroup.Id,
			})
			return originGroup.Id, nil
		}
	}

	if err := iterator.Error(); err != nil {
		return 0, fmt.Errorf("error iterating origin groups: %w", err)
	}

	return 0, fmt.Errorf("origin group with name %q not found in folder %s", name, folderID)
}
