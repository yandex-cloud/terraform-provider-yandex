package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func dataSourceYandexVPCSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCSecurityGroupRead,
		Schema: map[string]*schema.Schema{
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
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
				Elem:     aaa(),
				Set:      resourceYandexVPCSecurityGroupRuleHash,
			},

			"egress": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     aaa(),
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

func aaa() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"direction": {
				Type:     schema.TypeString,
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
			"protocol_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_number": {
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
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
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

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	securityGroup, err := config.sdk.VPC().SecurityGroup().Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: d.Get("security_group_id").(string),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Security group %q", d.Get("name").(string)))
	}

	createdAt, err := getTimestamp(securityGroup.GetCreatedAt())
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", securityGroup.GetName())
	d.Set("folder_id", securityGroup.GetFolderId())
	d.Set("network_id", securityGroup.GetNetworkId())
	d.Set("description", securityGroup.GetDescription())
	d.Set("status", securityGroup.GetStatus())

	ingress, egress := flattenSecurityGroupRulesSpec(securityGroup.Rules)

	if err := d.Set("ingress", ingress); err != nil {
		return err
	}
	if err := d.Set("egress", egress); err != nil {
		return err
	}

	return d.Set("labels", securityGroup.GetLabels())
}
