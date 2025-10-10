package cdn_rule

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	cdn_resource "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
)

func DataSourceCDNRuleSchema() schema.Schema {
	return schema.Schema{
		Description: "Get information about a Yandex CDN Resource Rule. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/).\n\n" +
			"~> **Note:** One of `rule_id` or `name` should be specified.",
		MarkdownDescription: "Get information about a Yandex CDN Resource Rule. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/).\n\n" +
			"~> **Note:** One of `rule_id` or `name` should be specified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the CDN rule in format `resource_id/rule_id`.",
				MarkdownDescription: "The ID of the CDN rule in format `resource_id/rule_id`.",
				Computed:            true,
			},
			"resource_id": schema.StringAttribute{
				Description:         "The ID of the CDN resource this rule belongs to.",
				MarkdownDescription: "The ID of the CDN resource this rule belongs to.",
				Required:            true,
			},
			"rule_id": schema.StringAttribute{
				Description:         "The ID of a specific CDN rule.",
				MarkdownDescription: "The ID of a specific CDN rule.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Description:         "The name of the CDN rule to search for.",
				MarkdownDescription: "The name of the CDN rule to search for.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("rule_id")),
				},
			},
			"rule_pattern": schema.StringAttribute{
				Description:         "Request path pattern for the rule (regular expression).",
				MarkdownDescription: "Request path pattern for the rule (regular expression).",
				Computed:            true,
			},
			"weight": schema.Int64Attribute{
				Description:         "Rule weight for determining priority. Higher weight means higher priority.",
				MarkdownDescription: "Rule weight for determining priority. Higher weight means higher priority.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"options": cdn_resource.CDNOptionsDataSourceSchema(),
		},
	}
}
