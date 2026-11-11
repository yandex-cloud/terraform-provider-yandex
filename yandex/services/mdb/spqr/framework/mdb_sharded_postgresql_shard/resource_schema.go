package mdb_sharded_postgresql_shard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func ShardSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Sharded PostgreSQL shard within the Yandex.Cloud",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Sharded PostgreSQL cluster. Provided by the client when the shard is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Sharded PostgreSQL shard. Provided by the client when the shard is added.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"shard_spec": schema.SingleNestedAttribute{
				MarkdownDescription: "Shard specification required to add shard into cluster.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"mdb_postgresql": schema.StringAttribute{
						Description: "ID of the Managed PostgreSQL cluster in Yandex Cloud. Provided by the client when the shard is added.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Validators: []validator.Object{
					//objectvalidator.ExactlyOneOf(
					//	path.MatchRoot("mdb_postgresql"),
					//),
				},
			},
		},
	}
}
