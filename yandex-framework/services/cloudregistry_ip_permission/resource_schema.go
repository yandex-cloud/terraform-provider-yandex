package cloudregistry_ip_permission

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func YandexCloudregistryIPPermissionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Creates a new Cloud Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/cloud-registry/operations/registry/registry-access)",
		MarkdownDescription: "Creates a new Cloud Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/cloud-registry/operations/registry/registry-access)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of IP permission.",
				Description:         "The ID of IP permission.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"registry_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the registry that IP restrictions are applied to.",
				Description:         "The ID of the registry that IP restrictions are applied to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"push": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of configured CIDRs from which `push` operations are allowed.",
				Description:         "List of configured CIDRs from which `push` operations are allowed.",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}(/\d{1,2})?$`),
							"must be a valid CIDR block",
						),
					),
					setvalidator.AtLeastOneOf(
						path.MatchRoot("pull"),
						path.MatchRoot("push"),
					),
				},
			},

			"pull": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of configured CIDRs from which `pull` operations are allowed.",
				Description:         "List of configured CIDRs from which `pull` operations are allowed.",
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}(/\d{1,2})?$`),
							"must be a valid CIDR block",
						),
					),
					setvalidator.AtLeastOneOf(
						path.MatchRoot("pull"),
						path.MatchRoot("push"),
					),
				},
			},

			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
