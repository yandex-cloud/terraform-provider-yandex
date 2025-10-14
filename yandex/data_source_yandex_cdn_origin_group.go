package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexCDNOriginGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex CDN Origin Group. For more information, see [the official documentation](https://yandex.cloud/docs/cdn/concepts/origins).\n\n~> One of `origin_group_id` or `name` should be specified.\n",
		Read:        dataSourceYandexCDNOriginGroupRead,
		Schema: map[string]*schema.Schema{
			"origin_group_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific origin group.",
				Computed:    true,
				Optional:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"provider_type": {
				Type:        schema.TypeString,
				Description: "CDN provider is a content delivery service provider",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
				Optional:    true,
			},
			"use_next": {
				Type:        schema.TypeBool,
				Description: resourceYandexCDNOriginGroup().Schema["use_next"].Description,
				Computed:    true,
			},
			"origin": {
				Type:        schema.TypeSet,
				Description: "A set of available origins.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:        schema.TypeString,
							Description: "IP address or Domain name of your origin and the port.",
							Required:    true,
						},
						"origin_group_id": {
							Type:        schema.TypeString,
							Description: "The ID of a specific origin group.",
							Computed:    true,
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "The origin is enabled and used as a source for the CDN.",
							Optional:    true,
							Default:     true,
						},
						"backup": {
							Type:        schema.TypeBool,
							Description: "Specifies whether the origin is used in its origin group as backup. A backup origin is used when one of active origins becomes unavailable.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexCDNOriginGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Reading CDN Origin Group %q", d.Id())

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("error getting folder ID while reading CDN origin group: %s", err)
	}

	var originGroupID int64

	if _, originGroupNameOk := d.GetOk("name"); originGroupNameOk {
		var (
			err error
		)

		groupName := d.Get("name").(string)
		originGroupID, err = resolveCDNOriginGroupID(ctx, config, folderID, groupName)
		if err != nil {
			return fmt.Errorf("failed to resolve data source cdn origin group by name: %v", err)
		}
	} else {
		originGroupID, _ = strconv.ParseInt(d.Get("origin_group_id").(string), 10, 64)
	}

	request, err := func() (*cdn.GetOriginGroupRequest, error) {
		return &cdn.GetOriginGroupRequest{
			FolderId:      folderID,
			OriginGroupId: originGroupID,
		}, nil
	}()

	if err != nil {
		return err
	}

	originGroup, err := config.sdk.CDN().OriginGroup().Get(ctx, request)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("origin group %q", d.Id()))
	}

	log.Printf("[DEBUG] Completed Reading CDN Origin Group %q", d.Id())

	if err := flattenYandexCDNOriginGroup(d, originGroup); err != nil {
		return err
	}

	return nil
}

func resolveCDNOriginGroupID(ctx context.Context, config *Config, folderID, name string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("empty name for origin group")
	}

	iterator := config.sdk.CDN().OriginGroup().OriginGroupIterator(ctx, &cdn.ListOriginGroupsRequest{
		FolderId: folderID,
	})

	for iterator.Next() {
		originGroup := iterator.Value()
		if name == originGroup.Name {
			return originGroup.Id, nil
		}
	}

	return 0, fmt.Errorf("origin name %q not found", name)
}
