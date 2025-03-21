package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexLoggingGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Logging group. For more information, see [the official documentation](https://yandex.cloud/docs/logging/concepts/log-group).\n\n~> If `group_id` is not specified `name` and `folder_id` will be used to designate Yandex Cloud Logging group.\n",

		Read: dataSourceYandexLoggingGroupRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeString,
				Description: "The Yandex Cloud Logging group ID.",
				Optional:    true,
				Computed:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},

			"retention_period": {
				Type:        schema.TypeString,
				Description: resourceYandexLoggingGroup().Schema["retention_period"].Description,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"data_stream": {
				Type:        schema.TypeString,
				Description: resourceYandexLoggingGroup().Schema["data_stream"].Description,
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["cloud_id"],
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexLoggingGroup().Schema["status"].Description,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexLoggingGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "group_id", "name")
	if err != nil {
		return err
	}

	groupID := d.Get("group_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		groupID, err = resolveObjectID(ctx, config, d, sdkresolvers.LogGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud Logging group by name: %v", err)
		}
	}

	req := logging.GetLogGroupRequest{
		LogGroupId: groupID,
	}

	group, err := config.sdk.Logging().LogGroup().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Logging group %q", d.Id()))
	}

	d.SetId(group.Id)
	d.Set("group_id", group.Id)
	return flattenYandexLoggingGroup(d, group)
}
