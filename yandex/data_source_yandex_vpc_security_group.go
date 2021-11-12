package yandex

import (
	"fmt"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexVPCSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCSecurityGroupRead,
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"ingress": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     dataSourceYandexSecurityGroupRule(),
				Set:      resourceYandexVPCSecurityGroupRuleHash,
			},

			"egress": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     dataSourceYandexSecurityGroupRule(),
				Set:      resourceYandexVPCSecurityGroupRuleHash,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"from_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"to_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"v4_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"predefined_target": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
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
