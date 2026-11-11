package pluginsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func DataSourceYandexComputePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Compute Placement group. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/placement-groups).\n\n~> One of `group_id` or `name` should be specified.\n",
		Read:        dataSourceYandexComputePlacementGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
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
			"placement_strategy": {
				Type:        schema.TypeMap,
				Description: "Placement strategy set for group.",
				Optional:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}

}

func dataSourceYandexComputePlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*providerconfig.Config)
	ctx := config.Context()

	err := sdkcommon.CheckOneOf(d, "group_id", "name")
	if err != nil {
		return err
	}

	groupID := d.Get("group_id").(string)
	_, groupNameOk := d.GetOk("name")

	if groupNameOk {
		groupID, err = sdkcommon.ResolveObjectIDV2(ctx, config, d, sdkcommon.ResolverWithClient(computesdk.NewPlacementGroupClient(config.SDK), computesdk.PlacementGroupResolver))
		if err != nil {
			return fmt.Errorf("failed to resolve data source Placement Group by name: %v", err)
		}
	}

	group, err := computesdk.NewPlacementGroupClient(config.SDK).Get(ctx, &compute.GetPlacementGroupRequest{
		PlacementGroupId: groupID,
	})

	if err != nil {
		return sdkcommon.HandleNotFoundError(err, d, fmt.Sprintf("snapshot with ID %q", groupID))
	}

	d.Set("group_id", group.Id)
	d.Set("folder_id", group.FolderId)
	d.Set("created_at", sdkcommon.GetTimestamp(group.CreatedAt))
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("placement_strategy", group.PlacementStrategy)
	if err := d.Set("labels", group.Labels); err != nil {
		return err
	}

	d.SetId(group.Id)

	return nil
}
