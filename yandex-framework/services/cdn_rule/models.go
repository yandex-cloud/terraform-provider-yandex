package cdn_rule

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CDNRuleModel represents the Terraform resource model for yandex_cdn_rule
type CDNRuleModel struct {
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	ID          types.String   `tfsdk:"id"`           // Composite: "resource_id/rule_id"
	ResourceID  types.String   `tfsdk:"resource_id"`  // CDN resource ID this rule belongs to
	RuleID      types.String   `tfsdk:"rule_id"`      // Rule ID (computed)
	Name        types.String   `tfsdk:"name"`         // Rule name
	RulePattern types.String   `tfsdk:"rule_pattern"` // Regular expression pattern
	Weight      types.Int64    `tfsdk:"weight"`       // Rule weight for ordering
	Options     types.List     `tfsdk:"options"`      // CDN options - uses same structure as cdn_resource
}

// CDNRuleDataSource represents the Terraform data source model for yandex_cdn_rule
type CDNRuleDataSource struct {
	ID          types.String `tfsdk:"id"`           // Composite: "resource_id/rule_id" (computed)
	ResourceID  types.String `tfsdk:"resource_id"`  // CDN resource ID (required)
	RuleID      types.String `tfsdk:"rule_id"`      // Rule ID for direct lookup (optional, computed) - String to match resource
	Name        types.String `tfsdk:"name"`         // Rule name for search (optional, computed)
	RulePattern types.String `tfsdk:"rule_pattern"` // Regular expression pattern (computed)
	Weight      types.Int64  `tfsdk:"weight"`       // Rule weight for ordering (computed)
	Options     types.List   `tfsdk:"options"`      // CDN options (computed)
}

// Note: Options uses CDNOptionsModel from cdn_resource package
// This ensures consistency between cdn_resource and cdn_rule
