package yandex

import (
	"fmt"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexVPCSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC Security Group Rule. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/security-groups).\n\nThis data source used to define Security Group Rule that can be used by other resources.\n",

		Read: dataSourceYandexVPCSecurityGroupRead,
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:        schema.TypeString,
				Description: "ID of Security Group that owns the rule.",
				Optional:    true,
				Computed:    true,
			},

			"network_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCSecurityGroup().Schema["network_id"].Description,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"ingress": {
				Type:        schema.TypeSet,
				Description: resourceYandexVPCSecurityGroup().Schema["ingress"].Description,
				Computed:    true,
				Elem:        dataSourceYandexSecurityGroupRule(),
				Set:         resourceYandexVPCSecurityGroupRuleHash,
			},

			"egress": {
				Type:        schema.TypeSet,
				Description: resourceYandexVPCSecurityGroup().Schema["egress"].Description,
				Computed:    true,
				Elem:        dataSourceYandexSecurityGroupRule(),
				Set:         resourceYandexVPCSecurityGroupRuleHash,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCSecurityGroup().Schema["status"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Description: resourceYandexSecurityGroupRule().Schema["description"].Description,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: resourceYandexSecurityGroupRule().Schema["labels"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"protocol": {
				Type:        schema.TypeString,
				Description: resourceYandexSecurityGroupRule().Schema["protocol"].Description,
				Computed:    true,
			},
			"port": {
				Type:        schema.TypeInt,
				Description: resourceYandexSecurityGroupRule().Schema["port"].Description,
				Computed:    true,
			},
			"from_port": {
				Type:        schema.TypeInt,
				Description: resourceYandexSecurityGroupRule().Schema["from_port"].Description,
				Computed:    true,
			},
			"to_port": {
				Type:        schema.TypeInt,
				Description: resourceYandexSecurityGroupRule().Schema["to_port"].Description,
				Computed:    true,
			},
			"v4_cidr_blocks": {
				Type:        schema.TypeList,
				Description: resourceYandexSecurityGroupRule().Schema["v4_cidr_blocks"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"v6_cidr_blocks": {
				Type:        schema.TypeList,
				Description: resourceYandexSecurityGroupRule().Schema["v6_cidr_blocks"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Description: resourceYandexSecurityGroupRule().Schema["security_group_id"].Description,
				Computed:    true,
			},
			"predefined_target": {
				Type:        schema.TypeString,
				Description: resourceYandexSecurityGroupRule().Schema["predefined_target"].Description,
				Computed:    true,
			},

			"id": {
				Type:        schema.TypeString,
				Description: resourceYandexSecurityGroupRule().Schema["id"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexVPCSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := checkOneOf(d, "security_group_id", "name")
	if err != nil {
		return err
	}

	sgID := d.Get("security_group_id").(string)
	_, nameOk := d.GetOk("name")

	if nameOk {
		sgID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.SecurityGroupResolver)
		if err != nil {
			return fmt.Errorf("VPC security group: failed to resolve data source security group by name: %v", err)
		}
	}

	if err := yandexVPCSecurityGroupRead(d, meta, sgID); err != nil {
		return err
	}

	d.SetId(sgID)

	return d.Set("security_group_id", sgID)
}
