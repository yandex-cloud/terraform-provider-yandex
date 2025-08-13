package cloudregistry_ip_permission

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func YandexCloudregistryIPPermissionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Creates a new Cloud Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/cloud-registry/operations/registry/registry-access)",
		MarkdownDescription: "Creates a new Cloud Registry IP Permission. For more information, see [the official documentation](https://yandex.cloud/docs/cloud-registry/operations/registry/registry-access)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of IP permission.",
				Description:         "The ID of IP permission.",
				Computed:            true,
			},

			"registry_name": schema.StringAttribute{
				MarkdownDescription: "The name of the registry that IP restrictions are applied to.",
				Description:         "The name of the registry that IP restrictions are applied to.",
				Optional:            true,
			},

			"registry_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the registry that IP restrictions are applied to.",
				Description:         "The ID of the registry that IP restrictions are applied to.",
				Optional:            true,
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
				},
			},

			"timeouts": timeouts.AttributesAll(ctx),
		},
		Blocks: map[string]schema.Block{},
	}
}
