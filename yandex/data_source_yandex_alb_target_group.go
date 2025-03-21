package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexALBTargetGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Application Load Balancer target group. For more information, see [Yandex Cloud Application Load Balancer](https://yandex.cloud/docs/application-load-balancer/quickstart).\n\nThis data source is used to define [Application Load Balancer Target Groups](https://yandex.cloud/docs/application-load-balancer/concepts/target-group) that can be used by other resources.\n\n~> One of `target_group_id` or `name` should be specified.\n",
		Read:        dataSourceYandexALBTargetGroupRead,
		Schema: map[string]*schema.Schema{
			"target_group_id": {
				Type:        schema.TypeString,
				Description: "Target Group ID.",
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
				Optional:    true,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"target": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Required: true,
						},
						"private_ipv4_address": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexALBTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "target_group_id", "name")
	if err != nil {
		return err
	}

	tgID := d.Get("target_group_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		tgID, err = resolveObjectID(ctx, config, d, sdkresolvers.ALBTargetGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source target group by name: %v", err)
		}
	}

	tg, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(ctx, &apploadbalancer.GetTargetGroupRequest{
		TargetGroupId: tgID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("target group with ID %q", tgID))
	}

	targets := flattenALBTargets(tg)

	d.Set("target_group_id", tg.Id)
	d.Set("name", tg.Name)
	d.Set("description", tg.Description)
	d.Set("created_at", getTimestamp(tg.CreatedAt))
	d.Set("folder_id", tg.FolderId)

	if err := d.Set("labels", tg.Labels); err != nil {
		return err
	}

	if err := d.Set("target", targets); err != nil {
		return err
	}

	d.SetId(tg.Id)

	return nil
}
