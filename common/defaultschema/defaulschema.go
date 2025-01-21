package defaultschema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func Id() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["id"],
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func FolderId() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["folder_id"],
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	}
}

func Name() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["name"],
		Optional:            true,
	}
}

func Description() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["description"],
		Optional:            true,
	}
}

func Labels() *schema.MapAttribute {
	return &schema.MapAttribute{
		MarkdownDescription: common.ResourceDescriptions["labels"],
		Optional:            true,
		ElementType:         types.StringType,
	}
}

func CreatedAt() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["created_at"],
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func DeletionProtection() *schema.BoolAttribute {
	return &schema.BoolAttribute{
		MarkdownDescription: common.ResourceDescriptions["deletion_protection"],
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
	}
}

func SecurityGroupIds() *schema.SetAttribute {
	return &schema.SetAttribute{
		MarkdownDescription: common.ResourceDescriptions["security_group_ids"],
		Optional:            true,
		ElementType:         types.StringType,
	}
}

func ServiceAccountId() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["service_account_id"],
		Required:            true,
	}
}

func SubnetIds() *schema.SetAttribute {
	return &schema.SetAttribute{
		MarkdownDescription: common.ResourceDescriptions["subnet_ids"],
		Required:            true,
		ElementType:         types.StringType,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.RequiresReplace(),
		},
	}
}

func NetworkId() *schema.StringAttribute {
	return &schema.StringAttribute{
		MarkdownDescription: common.ResourceDescriptions["network_id"],
		Required:            true,
	}
}
