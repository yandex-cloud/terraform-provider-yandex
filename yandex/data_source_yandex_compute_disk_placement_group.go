package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexComputeDiskPlacementGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Compute Disk Placement group. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/disk#nr-disks).\n\n~> One of `group_id` or `name` should be specified.\n",

		Read: dataSourceYandexComputeDiskPlacementGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: common.ResourceDescriptions["name"],
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"group_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific group.",
				Optional:    true,
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Optional:    true,
				Default:     "ru-central1-b",
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexComputeDiskPlacementGroup().Schema["status"].Description,
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

func dataSourceYandexComputeDiskPlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "group_id", "name")
	if err != nil {
		return err
	}

	groupID := d.Get("group_id").(string)
	_, groupNameOk := d.GetOk("name")

	if groupNameOk {
		groupID, err = resolveObjectID(ctx, config, d, sdkresolvers.DiskPlacementGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Placement Group by name: %v", err)
		}
	}

	group, err := config.sdk.Compute().DiskPlacementGroup().Get(ctx, &compute.GetDiskPlacementGroupRequest{
		DiskPlacementGroupId: groupID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("snapshot with ID %q", groupID))
	}

	d.Set("group_id", group.Id)
	d.Set("folder_id", group.FolderId)
	d.Set("created_at", getTimestamp(group.CreatedAt))
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("zone", group.ZoneId)
	d.Set("status", group.Status.String())

	if err := d.Set("labels", group.Labels); err != nil {
		return err
	}

	d.SetId(group.Id)

	return nil
}
