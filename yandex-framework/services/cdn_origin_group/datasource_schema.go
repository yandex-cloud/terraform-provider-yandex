package cdn_origin_group

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func DataSourceCDNOriginGroupSchema() schema.Schema {
	return schema.Schema{
		Description: "Get information about a Yandex CDN Origin Group. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/origins).\n\n" +
			"~> **Note:** One of `origin_group_id` or `name` should be specified.",
		MarkdownDescription: "Get information about a Yandex CDN Origin Group. For more information, see " +
			"[the official documentation](https://yandex.cloud/docs/cdn/concepts/origins).\n\n" +
			"~> **Note:** One of `origin_group_id` or `name` should be specified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The ID of the CDN origin group (stored as string).",
				MarkdownDescription: "The ID of the CDN origin group (stored as string).",
				Computed:            true,
			},
			"origin_group_id": schema.StringAttribute{
				Description:         "The ID of a specific origin group.",
				MarkdownDescription: "The ID of a specific origin group.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("name")),
				},
			},
			"folder_id": schema.StringAttribute{
				Description:         common.ResourceDescriptions["folder_id"],
				MarkdownDescription: common.ResourceDescriptions["folder_id"],
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Description:         common.ResourceDescriptions["name"],
				MarkdownDescription: common.ResourceDescriptions["name"],
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("origin_group_id")),
				},
			},
			"provider_type": schema.StringAttribute{
				Description:         "CDN provider type.",
				MarkdownDescription: "CDN provider type.",
				Computed:            true,
			},
			"use_next": schema.BoolAttribute{
				Description: "If true, the next origin in group will be used if current origin fails. " +
					"If false, the request will fail.",
				MarkdownDescription: "If `true`, the next origin in group will be used if current origin fails. " +
					"If `false`, the request will fail.",
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"origin": schema.SetNestedBlock{
				Description:         "A set of available origins in the group.",
				MarkdownDescription: "A set of available origins in the group.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description:         "IP address or domain name of your origin and the port (e.g., example.com:8080).",
							MarkdownDescription: "IP address or domain name of your origin and the port (e.g., `example.com:8080`).",
							Computed:            true,
						},
						"origin_group_id": schema.StringAttribute{
							Description:         "The ID of the origin group this origin belongs to.",
							MarkdownDescription: "The ID of the origin group this origin belongs to.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							Description:         "Whether the origin is enabled and used as a source for the CDN.",
							MarkdownDescription: "Whether the origin is enabled and used as a source for the CDN.",
							Computed:            true,
						},
						"backup": schema.BoolAttribute{
							Description: "Specifies whether the origin is used in its origin group as backup. " +
								"A backup origin is used when one of active origins becomes unavailable.",
							MarkdownDescription: "Specifies whether the origin is used in its origin group as backup. " +
								"A backup origin is used when one of active origins becomes unavailable.",
							Computed: true,
						},
					},
				},
			},
		},
	}
}
