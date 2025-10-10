package cdn_origin_group

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CDNOriginGroupModel represents the Terraform resource model for yandex_cdn_origin_group
type CDNOriginGroupModel struct {
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
	ID           types.String   `tfsdk:"id"`
	FolderID     types.String   `tfsdk:"folder_id"`
	Name         types.String   `tfsdk:"name"`
	ProviderType types.String   `tfsdk:"provider_type"`
	UseNext      types.Bool     `tfsdk:"use_next"`
	Origins      types.Set      `tfsdk:"origin"`
}

// CDNOriginGroupDataSource represents the Terraform data source model for yandex_cdn_origin_group
type CDNOriginGroupDataSource struct {
	ID            types.String `tfsdk:"id"`
	OriginGroupID types.Int64  `tfsdk:"origin_group_id"`
	FolderID      types.String `tfsdk:"folder_id"`
	Name          types.String `tfsdk:"name"`
	ProviderType  types.String `tfsdk:"provider_type"`
	UseNext       types.Bool   `tfsdk:"use_next"`
	Origins       types.Set    `tfsdk:"origin"`
}

// OriginModel represents a single origin in the origin group
type OriginModel struct {
	Source        types.String `tfsdk:"source"`
	OriginGroupID types.Int64  `tfsdk:"origin_group_id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Backup        types.Bool   `tfsdk:"backup"`
}
