package mdb_sharded_postgresql_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func UserSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Sharded PostgreSQL user within the Yandex.Cloud",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Sharded PostgreSQL cluster. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Sharded PostgreSQL user. Provided by the client when the user is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the Sharded PostgreSQL user. Provided by the client when the user is created.",
				Optional:            true,
				Sensitive:           true,
			},
			"settings": SettingsSchema(),
			"grants":   GrantsSchema(),
		},
		Blocks: map[string]schema.Block{
			"permissions": PermissionSchema(),
		},
	}
}

func PermissionSchema() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "Block represents databases that are permitted to user.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database": schema.StringAttribute{
					MarkdownDescription: "Name of the database that the permission grants access to.",
					Required:            true,
				},
			},
		},
	}
}

func GrantsSchema() schema.SetAttribute {
	return schema.SetAttribute{
		MarkdownDescription: "",
		ElementType:         types.StringType,
		Optional:            true,
	}
}

func SettingsSchema() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "",
		CustomType:          mdbcommon.NewSettingsMapType(attrProvider),
		ElementType:         types.StringType,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.UseStateForUnknown(),
		},
	}
}
