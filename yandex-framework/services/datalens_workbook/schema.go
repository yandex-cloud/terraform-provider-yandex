package datalens_workbook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

func ResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a DataLens workbook resource. Workbooks are top-level containers for " +
			"DataLens connections, datasets, charts and dashboards. " +
			"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start).",
		Attributes: map[string]schema.Attribute{
			"id": defaultschema.Id(),
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID for the DataLens instance. " +
					"If not specified, the provider-level `organization_id` is used.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collection_id": schema.StringAttribute{
				MarkdownDescription: "The parent collection ID. If unset, the workbook is created at the organization root. " +
					"Changing this attribute forces recreation of the resource.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The workbook title.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The workbook description.",
				Optional:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The DataLens tenant ID the workbook belongs to.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The workbook lifecycle status. One of `creating`, `active`, `deleting`.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "Identity of the user who created the workbook.",
				Computed:            true,
			},
			"created_at": defaultschema.CreatedAt(),
			"updated_by": schema.StringAttribute{
				MarkdownDescription: "Identity of the user who last updated the workbook.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The last update timestamp of the resource.",
				Computed:            true,
			},
		},
	}
}
