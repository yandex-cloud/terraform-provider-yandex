package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NodeResource() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Validators: []validator.Object{
			objectvalidator.IsRequired(),
		},
		Attributes: NodeResourceAttributes(),
	}
}

func NodeResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"resource_preset_id": schema.StringAttribute{Required: true},
		"disk_size":          schema.Int64Attribute{Required: true},
		"disk_type_id":       schema.StringAttribute{Required: true},
	}
}

func Hosts() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:    true,
		Description: "Current nodes in the cluster",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"fqdn": schema.StringAttribute{Computed: true},
				"zone": schema.StringAttribute{Computed: true},
				"type": schema.StringAttribute{Computed: true},
				"roles": schema.SetAttribute{
					Computed:    true,
					ElementType: types.StringType,
				},
				"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
				"subnet_id":        schema.StringAttribute{Computed: true, Optional: true},
				"node_group":       schema.StringAttribute{Computed: true},
			},
		},
	}
}
