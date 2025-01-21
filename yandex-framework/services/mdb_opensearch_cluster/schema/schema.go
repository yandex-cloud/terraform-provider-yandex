package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NodeResource() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Resources allocated to hosts of this OpenSearch node group.",
		Validators: []validator.Object{
			objectvalidator.IsRequired(),
		},
		Attributes: NodeResourceAttributes(),
	}
}

func NodeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"resource_preset_id": schema.StringAttribute{
			MarkdownDescription: "The ID of the preset for computational resources available to a host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-opensearch/concepts).",
			Required:            true,
		},
		"disk_size": schema.Int64Attribute{
			MarkdownDescription: "Volume of the storage available to a host, in bytes.",
			Required:            true,
		},
		"disk_type_id": schema.StringAttribute{
			MarkdownDescription: "Type of the storage of OpenSearch hosts.",
			Required:            true,
		},
	}
}

func Hosts() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:            true,
		MarkdownDescription: "A hosts of the OpenSearch cluster.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"fqdn": schema.StringAttribute{
					MarkdownDescription: "The fully qualified domain name of the host.",
					Computed:            true,
				},
				"zone": schema.StringAttribute{
					MarkdownDescription: "The availability zone where the OpenSearch host will be created. For more information see [the official documentation](https://yandex.cloud/docs/overview/concepts/geo-scope).",
					Computed:            true,
				},
				"type": schema.StringAttribute{
					MarkdownDescription: "The type of the deployed host. Can be either `OPENSEARCH` or `DASHBOARDS`.",
					Computed:            true,
				},
				"roles": schema.SetAttribute{
					MarkdownDescription: "The roles of the deployed host. Can contain `DATA` and/or `MANAGER` roles. Will be empty for `DASHBOARDS` type.",
					Computed:            true,
					ElementType:         types.StringType,
				},
				"assign_public_ip": schema.BoolAttribute{
					MarkdownDescription: "Sets whether the host should get a public IP address. Can be either `true` or `false`.",
					Computed:            true,
					Optional:            true,
				},
				"subnet_id": schema.StringAttribute{
					MarkdownDescription: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
					Computed:            true,
					Optional:            true,
				},
				"node_group": schema.StringAttribute{
					MarkdownDescription: "Name of the node group.",
					Computed:            true,
				},
			},
		},
	}
}
