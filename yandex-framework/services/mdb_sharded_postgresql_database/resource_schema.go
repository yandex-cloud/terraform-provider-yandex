package mdb_sharded_postgresql_database

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func DatabaseSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Sharded PostgreSQL database within the Yandex.Cloud",
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
		},
	}
}
