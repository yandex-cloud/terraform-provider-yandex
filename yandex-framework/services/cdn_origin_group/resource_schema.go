package cdn_origin_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func CDNOriginGroupSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Allows management of [Yandex Cloud CDN Origin Groups](https://yandex.cloud/docs/cdn/concepts/origins).\n\n" +
			"~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: " +
			"`yc cdn provider activate --folder-id <folder-id> --type gcore`.",
		MarkdownDescription: "Allows management of [Yandex Cloud CDN Origin Groups](https://yandex.cloud/docs/cdn/concepts/origins).\n\n" +
			"~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: " +
			"`yc cdn provider activate --folder-id <folder-id> --type gcore`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the CDN origin group.",
				MarkdownDescription: "The ID of the CDN origin group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.",
				MarkdownDescription: "The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the origin group.",
				MarkdownDescription: "Name of the origin group.",
			},
			"provider_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         `CDN provider is a content delivery service provider. Possible values: "ourcdn" (default) or "gcore".`,
				MarkdownDescription: `CDN provider is a content delivery service provider. Possible values: "ourcdn" (default) or "gcore".`,
				Validators: []validator.String{
					stringvalidator.OneOf("ourcdn", "gcore"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_next": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "If the option is active (has true value), in case the origin responds with 4XX or 5XX codes, use the next origin from the list.",
				MarkdownDescription: "If the option is active (has true value), in case the origin responds with 4XX or 5XX codes, use the next origin from the list.",
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),

			"origin": schema.SetNestedBlock{
				Description:         "A set of available origins. An origin group must contain at least one enabled origin.",
				MarkdownDescription: "A set of available origins. An origin group must contain at least one enabled origin.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Required:            true,
							Description:         "IP address or Domain name of your origin and the port (e.g., example.com:8080 or 192.0.2.1:80).",
							MarkdownDescription: "IP address or Domain name of your origin and the port (e.g., `example.com:8080` or `192.0.2.1:80`).",
						},
						"origin_group_id": schema.Int64Attribute{
							Computed:            true,
							Description:         "The ID of the origin group that this origin belongs to.",
							MarkdownDescription: "The ID of the origin group that this origin belongs to.",
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"enabled": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
							Description:         "The origin is enabled and used as a source for the CDN. Default: true.",
							MarkdownDescription: "The origin is enabled and used as a source for the CDN. Default: `true`.",
						},
						"backup": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							Description:         "Specifies whether the origin is used in its origin group as backup. A backup origin is used when one of active origins becomes unavailable. Default: false.",
							MarkdownDescription: "Specifies whether the origin is used in its origin group as backup. A backup origin is used when one of active origins becomes unavailable. Default: `false`.",
						},
					},
				},
			},
		},
	}
}
