package yandex

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

const (
	yandexCDNOriginGroupDefaultTimeout = 2 * time.Minute
)

func resourceYandexCDNOriginGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexCDNOriginGroupCreate,
		Read:   resourceYandexCDNOriginGroupRead,
		Update: resourceYandexCDNOriginGroupUpdate,
		Delete: resourceYandexCDNOriginGroupDelete,

		SchemaVersion: 0,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexCDNOriginGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexCDNOriginGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexCDNOriginGroupDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"use_next": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"origin": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,

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

func expandCDNOriginParams(origin map[string]interface{}) *cdn.OriginParams {
	if origin == nil {
		return nil
	}

	originParams := &cdn.OriginParams{}

	if v, ok := origin["source"]; ok {
		originParams.Source = v.(string)
	}

	if v, ok := origin["enabled"]; ok {
		originParams.Enabled = v.(bool)
	}

	if v, ok := origin["backup"]; ok {
		originParams.Backup = v.(bool)
	}

	return originParams
}

func prepareCDNCreateOriginGroupRequest(d *schema.ResourceData, meta *Config) (*cdn.CreateOriginGroupRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
	}

	var useNext *wrappers.BoolValue
	if v, ok := d.GetOk("use_next"); ok {
		useNext = &wrappers.BoolValue{Value: v.(bool)}
	}

	log.Printf("[DEBUG] Preparing create CDN Origin Group request %q", d.Get("name").(string))

	result := &cdn.CreateOriginGroupRequest{
		FolderId: folderID,
		Name:     d.Get("name").(string),

		UseNext: useNext,
	}

	for _, origin := range d.Get("origin").(*schema.Set).List() {
		originMap, ok := origin.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type for CDN Origin Group: %T", origin)
		}

		result.Origins = append(result.Origins, expandCDNOriginParams(originMap))
	}

	return result, nil
}

func resourceYandexCDNOriginGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Crating CDN Origin Group %q", d.Get("name").(string))

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	request, err := prepareCDNCreateOriginGroupRequest(d, config)
	if err != nil {
		return err
	}

	operation, err := config.sdk.WrapOperation(config.sdk.CDN().OriginGroup().Create(ctx, request))
	if err != nil {
		return fmt.Errorf("error while requesting API to create CDN Origin Group: %s", err)
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadata for CDN Origin Group: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.CreateOriginGroupMetadata)
	if !ok {
		return fmt.Errorf("origin group metadata type mismatch on create")
	}

	d.SetId(strconv.FormatInt(pm.OriginGroupId, 10))

	err = operation.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while requesting API to create origin group: %s", err)
	}

	return resourceYandexCDNOriginGroupRead(d, meta)
}

func flattenYandexCDNOrigins(origins []*cdn.Origin) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(origins))

	for i := range origins {
		om := make(map[string]interface{})

		om["origin_group_id"] = origins[i].OriginGroupId
		om["source"] = origins[i].Source
		om["enabled"] = origins[i].Enabled
		om["backup"] = origins[i].Backup

		result = append(result, om)
	}

	return result
}

func flattenYandexCDNOriginGroup(d *schema.ResourceData, origin *cdn.OriginGroup) error {
	d.SetId(strconv.FormatInt(origin.Id, 10))

	d.Set("folder_id", origin.FolderId)
	d.Set("name", origin.Name)
	d.Set("use_next", origin.UseNext)

	if err := d.Set("origin", flattenYandexCDNOrigins(origin.Origins)); err != nil {
		return err
	}

	return nil
}

func resourceYandexCDNOriginGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Reading CDN Origin Group %q", d.Id())

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	request, err := func() (*cdn.GetOriginGroupRequest, error) {
		folderID, err := getFolderID(d, config)
		if err != nil {
			return nil, fmt.Errorf("error getting folder ID while creating instance: %s", err)
		}

		groupID, err := strconv.ParseInt(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}

		return &cdn.GetOriginGroupRequest{
			FolderId:      folderID,
			OriginGroupId: groupID,
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

	return flattenYandexCDNOriginGroup(d, originGroup)
}

func prepareCDNUpdateOriginGroupRequest(d *schema.ResourceData, config *Config) (*cdn.UpdateOriginGroupRequest, error) {
	folderID, err := getFolderID(d, config)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance: %s", err)
	}

	groupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, err
	}

	result := &cdn.UpdateOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: groupID,
	}

	if d.HasChange("name") {
		result.GroupName = &wrappers.StringValue{Value: d.Get("name").(string)}
	}

	if d.HasChange("use_next") {
		result.UseNext = &wrappers.BoolValue{Value: d.Get("use_next").(bool)}
	}

	for _, v := range d.Get("origin").(*schema.Set).List() {
		originParam := v.(map[string]interface{})

		result.Origins = append(result.Origins, &cdn.OriginParams{
			Source:  originParam["source"].(string),
			Enabled: originParam["enabled"].(bool),
			Backup:  originParam["backup"].(bool),
		})
	}

	return result, nil
}

func resourceYandexCDNOriginGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating CDN Origin Group %q", d.Id())

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// NOTE: Update has PUT semantics

	request, err := prepareCDNUpdateOriginGroupRequest(d, config)
	if err != nil {
		return err
	}

	operation, err := config.sdk.WrapOperation(config.sdk.CDN().OriginGroup().Update(ctx, request))
	if err != nil {
		return err
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadate for CDN Origin Group update: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.UpdateOriginGroupMetadata)
	if !ok {
		return fmt.Errorf("origin group metadata type mismatch on update")
	}

	d.SetId(strconv.FormatInt(pm.OriginGroupId, 10))

	err = operation.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while requesting API to update CDN Origin Group: %s", err)
	}

	log.Printf("[DEBUG] Completed updating CDN Origin Group %q", d.Id())

	return resourceYandexCDNOriginGroupRead(d, meta)
}

func prepareCDNDeleteOriginGroupRequest(d *schema.ResourceData, config *Config) (*cdn.DeleteOriginGroupRequest, error) {
	folderID, err := getFolderID(d, config)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while deleting cdn origin group: %s", err)
	}

	groupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return nil, err
	}

	result := &cdn.DeleteOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: groupID,
	}

	return result, nil
}

func resourceYandexCDNOriginGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting CDN Origin Group %q", d.Id())

	request, err := prepareCDNDeleteOriginGroupRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	operation, err := config.sdk.WrapOperation(config.sdk.CDN().OriginGroup().Delete(ctx, request))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Origin Group ID: %d", request.OriginGroupId))
	}

	protoMetadata, err := operation.Metadata()
	if err != nil {
		return fmt.Errorf("error while obtaining response metadata for CDN Origin Group: %s", err)
	}

	pm, ok := protoMetadata.(*cdn.DeleteOriginGroupMetadata)
	if !ok {
		return fmt.Errorf("origin group metadata type mismatch on delete")
	}

	log.Printf("[DEBUG] Waiting Deleting CDN Origin Group operation completion %q", d.Id())

	if err = operation.Wait(ctx); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Origin Group %q: %#v", d.Id(), pm.OriginGroupId)

	return nil
}
