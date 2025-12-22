package organizationmanager_mfa_enforcement_audience

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "MFA enforcement audience resource",
		Attributes: map[string]schema.Attribute{
			"mfa_enforcement_id": schema.StringAttribute{
				MarkdownDescription: "MFA enforcement id",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subject_id": schema.StringAttribute{
				MarkdownDescription: "Subject id",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
