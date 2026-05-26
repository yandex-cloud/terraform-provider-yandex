package validate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

// DiskSizeXorValidator ensures exactly one of disk_size or disk_size_gb is set in each
// resources block, and that exactly one of disk_size_limit or disk_size_gb_limit is set in each
// configured disk_size_autoscaling block.
type DiskSizeXorValidator struct{}

func (v DiskSizeXorValidator) Description(context.Context) string {
	return "Each node group `resources` block must set exactly one of `disk_size` or `disk_size_gb`, and each configured `disk_size_autoscaling` block must set exactly one of `disk_size_limit` or `disk_size_gb_limit`."
}

func (v DiskSizeXorValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v DiskSizeXorValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var root model.OpenSearch
	resp.Diagnostics.Append(req.Config.Get(ctx, &root)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if root.Config.IsNull() || root.Config.IsUnknown() {
		return
	}

	var cfg model.Config
	resp.Diagnostics.Append(root.Config.As(ctx, &cfg, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !(cfg.OpenSearch.IsNull() || cfg.OpenSearch.IsUnknown()) {
		var osCfg model.OpenSearchSubConfig
		resp.Diagnostics.Append(cfg.OpenSearch.As(ctx, &osCfg, datasize.DefaultOpts)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var groups []model.OpenSearchNode
		if !(osCfg.NodeGroups.IsNull() || osCfg.NodeGroups.IsUnknown()) {
			resp.Diagnostics.Append(osCfg.NodeGroups.ElementsAs(ctx, &groups, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			for i, ng := range groups {
				p := path.Root("config").AtName("opensearch").AtName("node_groups").AtListIndex(i).AtName("resources")
				validateResourcesDiskXor(ctx, ng.Resources, p, resp)

				ap := path.Root("config").AtName("opensearch").AtName("node_groups").AtListIndex(i).AtName("disk_size_autoscaling")
				validateAutoscalingDiskLimitXor(ctx, ng.DiskSizeAutoscaling, ap, resp)
			}
		}
	}

	if !(cfg.Dashboards.IsNull() || cfg.Dashboards.IsUnknown()) {
		var dashCfg model.DashboardsSubConfig
		resp.Diagnostics.Append(cfg.Dashboards.As(ctx, &dashCfg, datasize.DefaultOpts)...)
		if resp.Diagnostics.HasError() {
			return
		}
		var groups []model.DashboardNode
		if !(dashCfg.NodeGroups.IsNull() || dashCfg.NodeGroups.IsUnknown()) {
			resp.Diagnostics.Append(dashCfg.NodeGroups.ElementsAs(ctx, &groups, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			for i, ng := range groups {
				p := path.Root("config").AtName("dashboards").AtName("node_groups").AtListIndex(i).AtName("resources")
				validateResourcesDiskXor(ctx, ng.Resources, p, resp)
			}
		}
	}
}

func validateResourcesDiskXor(ctx context.Context, resources types.Object, base path.Path, resp *resource.ValidateConfigResponse) {
	if resources.IsNull() || resources.IsUnknown() {
		return
	}
	var nr model.NodeResource
	resp.Diagnostics.Append(resources.As(ctx, &nr, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasBytes := !nr.DiskSize.IsNull() && !nr.DiskSize.IsUnknown()
	hasGb := !nr.DiskSizeGb.IsNull() && !nr.DiskSizeGb.IsUnknown()

	switch {
	case hasBytes && hasGb:
		resp.Diagnostics.AddAttributeError(
			base,
			"Invalid disk size attributes",
			"Specify exactly one of `disk_size` (bytes) or `disk_size_gb` (GiB), not both.",
		)
	case !hasBytes && !hasGb:
		resp.Diagnostics.AddAttributeError(
			base,
			"Invalid disk size attributes",
			fmt.Sprintf("%s must set exactly one of `disk_size` or `disk_size_gb`.", base),
		)
	}
}

func validateAutoscalingDiskLimitXor(ctx context.Context, autoscaling types.Object, base path.Path, resp *resource.ValidateConfigResponse) {
	if autoscaling.IsNull() || autoscaling.IsUnknown() {
		return
	}
	var a model.DiskSizeAutoscaling
	resp.Diagnostics.Append(autoscaling.As(ctx, &a, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasBytes := !a.DiskSizeLimit.IsNull() && !a.DiskSizeLimit.IsUnknown()
	hasGb := !a.DiskSizeGbLimit.IsNull() && !a.DiskSizeGbLimit.IsUnknown()

	switch {
	case hasBytes && hasGb:
		resp.Diagnostics.AddAttributeError(
			base,
			"Invalid disk size autoscaling attributes",
			"Specify exactly one of `disk_size_limit` (bytes) or `disk_size_gb_limit` (GiB), not both.",
		)
	case !hasBytes && !hasGb:
		resp.Diagnostics.AddAttributeError(
			base,
			"Invalid disk size autoscaling attributes",
			fmt.Sprintf("%s must set exactly one of `disk_size_limit` or `disk_size_gb_limit`.", base),
		)
	}
}
