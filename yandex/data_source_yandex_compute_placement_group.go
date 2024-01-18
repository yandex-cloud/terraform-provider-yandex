package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputePlacementGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"placement_strategy": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceYandexComputePlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "group_id", "name")
	if err != nil {
		return err
	}

	groupID := d.Get("group_id").(string)
	_, groupNameOk := d.GetOk("name")

	if groupNameOk {
		groupID, err = resolveObjectID(ctx, config, d, sdkresolvers.PlacementGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Placement Group by name: %v", err)
		}
	}

	group, err := config.sdk.Compute().PlacementGroup().Get(ctx, &compute.GetPlacementGroupRequest{
		PlacementGroupId: groupID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("snapshot with ID %q", groupID))
	}

	d.Set("group_id", group.Id)
	d.Set("folder_id", group.FolderId)
	d.Set("created_at", getTimestamp(group.CreatedAt))
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("placement_strategy", group.PlacementStrategy)
	if err := d.Set("labels", group.Labels); err != nil {
		return err
	}

	d.SetId(group.Id)

	return nil
}
