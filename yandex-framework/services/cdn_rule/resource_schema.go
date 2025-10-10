package cdn_rule

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
)

func CDNRuleSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a CDN Rule. CDN Rules allow you to apply different CDN settings based on URL patterns.\n\n" +
			"Rules with lower weight values are executed first.",
		MarkdownDescription: "Manages a CDN Rule. CDN Rules allow you to apply different CDN settings based on URL patterns.\n\n" +
			"Rules with lower weight values are executed first.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the CDN rule in the format 'resource_id/rule_id'.",
				MarkdownDescription: "The ID of the CDN rule in the format `resource_id/rule_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_id": schema.StringAttribute{
				Required:            true,
				Description:         "CDN Resource ID to attach the rule to.",
				MarkdownDescription: "CDN Resource ID to attach the rule to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.StringAttribute{
				Computed:            true,
				Description:         "Rule ID (stored as string to avoid int64 precision loss).",
				MarkdownDescription: "Rule ID (stored as string to avoid int64 precision loss).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Rule name (max 50 characters).",
				MarkdownDescription: "Rule name (max 50 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
				},
			},
			"rule_pattern": schema.StringAttribute{
				Required:            true,
				Description:         "Rule pattern - must be a valid regular expression (max 100 characters).",
				MarkdownDescription: "Rule pattern - must be a valid regular expression (max 100 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^.+$`),
						"must be a valid regular expression",
					),
				},
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Description:         "Rule weight (0-9999) - rules with lower weights execute first. Default: 0.",
				MarkdownDescription: "Rule weight (0-9999) - rules with lower weights execute first. Default: `0`.",
				Validators: []validator.Int64{
					int64validator.Between(0, 9999),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),

			// Reuse options schema from cdn_resource - identical structure
			"options": cdn_resource.CDNOptionsSchema(),
		},
	}
}
