package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func dataSourceYandexCDNOriginGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexCDNOriginGroupRead,
		Schema: map[string]*schema.Schema{
			"origin_group_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"use_next": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeSet,
				Computed: true,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"origin_group_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"backup": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

	originGroupID := int64(d.Get("origin_group_id").(int))
	_, originGroupNameOk := d.GetOk("name")

	if originGroupNameOk {
		var (
			err error
		)

		groupName := d.Get("name").(string)
		originGroupID, err = resolveCDNOriginGroupID(ctx, config, folderID, groupName)
		if err != nil {
			return fmt.Errorf("failed to resolve data source cdn origin group by name: %v", err)
		}
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
