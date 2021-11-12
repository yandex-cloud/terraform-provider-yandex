package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexLBTargetGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLBTargetGroupRead,
		Schema: map[string]*schema.Schema{
			"target_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
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

			"target": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: resourceLBTargetGroupTargetHash,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexLBTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "target_group_id", "name")
	if err != nil {
		return err
	}

	tgID := d.Get("target_group_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		tgID, err = resolveObjectID(ctx, config, d, sdkresolvers.TargetGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source target group by name: %v", err)
		}
	}

	tg, err := config.sdk.LoadBalancer().TargetGroup().Get(ctx, &loadbalancer.GetTargetGroupRequest{
		TargetGroupId: tgID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("target group with ID %q", tgID))
	}

	targets, err := flattenLBTargets(tg)
	if err != nil {
		return err
	}

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
