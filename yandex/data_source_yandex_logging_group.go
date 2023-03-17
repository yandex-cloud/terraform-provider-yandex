package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexLoggingGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexLoggingGroupRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"group_id": {
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

			"retention_period": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"data_stream": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"cloud_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
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
