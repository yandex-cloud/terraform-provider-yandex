package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexVPCSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCSecurityGroupRuleRead,
		Schema: map[string]*schema.Schema{
			"rule_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_group_binding": {
				Type:     schema.TypeString,
				Required: true,
			},
			"direction": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"predefined_target": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexVPCSecurityGroupRuleRead(data *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ruleId := data.Get("rule_id").(string)
	sgId := data.Get("security_group_binding").(string)

	ctx, cancel := context.WithTimeout(config.Context(), yandexVPCSecurityGroupDefaultTimeout)
	defer cancel()

	rule, err := findRule(data, config, ctx, sgId, ruleId)
	if err != nil {
		return err
	}

	data.SetId(ruleId)

	return writeSecurityGroupRuleToData(rule, data)
}
