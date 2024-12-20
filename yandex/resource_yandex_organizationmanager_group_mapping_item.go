package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

const yandexOrganizationManagerGroupMappingItemDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerGroupMappingItem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerGroupMappingItemCreate,
		ReadContext:   resourceYandexOrganizationManagerGroupMappingItemRead,
		DeleteContext: resourceYandexOrganizationManagerGroupMappingItemDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerGroupMappingItemDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerGroupMappingItemDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerGroupMappingItemDefaultTimeout),
		},

		Importer: &schema.ResourceImporter{
			StateContext: groupMappingItemImportStateContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the SAML Federation",
			},
			"internal_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the internal group",
			},
			"external_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the external group",
			},
		},
	}
}

func resourceYandexOrganizationManagerGroupMappingItemCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID := d.Get("federation_id").(string)
	internalGroupID := d.Get("internal_group_id").(string)
	externalGroupID := d.Get("external_group_id").(string)

	err := updateGroupMappingItems(context, config, organizationmanager.GroupMappingItemDelta_ADD,
		federationID, internalGroupID, externalGroupID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strings.Join([]string{federationID, internalGroupID, externalGroupID}, "/"))

	return resourceYandexOrganizationManagerGroupMappingItemRead(context, d, meta)
}

func resourceYandexOrganizationManagerGroupMappingItemRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID := d.Get("federation_id").(string)

	items, err := getGroupMappingItems(context, config, federationID)

	if err != nil {
		return diag.FromErr(err)
	}

	internalGroupID := d.Get("internal_group_id").(string)
	externalGroupID := d.Get("external_group_id").(string)
	for _, item := range items {
		if item.InternalGroupId == internalGroupID &&
			item.ExternalGroupId == externalGroupID {
			return nil
		}
	}

	log.Printf("[DEBUG]: Group-mapping item with internal-id %s and external-id %s not found for federation %s, removing from state file.", internalGroupID, externalGroupID, federationID)
	d.SetId("")
	return nil
}

func resourceYandexOrganizationManagerGroupMappingItemDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID := d.Get("federation_id").(string)
	internalGroupID := d.Get("internal_group_id").(string)
	externalGroupID := d.Get("external_group_id").(string)

	err := updateGroupMappingItems(context, config, organizationmanager.GroupMappingItemDelta_REMOVE,
		federationID, internalGroupID, externalGroupID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateGroupMappingItems(ctx context.Context, config *Config, action organizationmanager.GroupMappingItemDelta_Action, federationID, internalID, externalID string) error {
	req := &organizationmanager.UpdateGroupMappingItemsRequest{
		FederationId: federationID,
		GroupMappingItemDeltas: []*organizationmanager.GroupMappingItemDelta{
			{
				Action: action,
				Item: &organizationmanager.GroupMappingItem{
					ExternalGroupId: externalID,
					InternalGroupId: internalID,
				},
			},
		},
	}
	log.Printf("[DEBUG] group-mapping update-items request: %s", protoDump(req))

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().GroupMapping().UpdateItems(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update group mapping items op %s: %s", action, err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to update group mapping items: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("group mapping update failed: %s", err)
	}

	return nil
}

func getGroupMappingItems(ctx context.Context, config *Config, federationID string) ([]*organizationmanager.GroupMappingItem, error) {
	items := make([]*organizationmanager.GroupMappingItem, 0)

	pageToken := ""
	for {
		req := &organizationmanager.ListGroupMappingItemsRequest{
			FederationId: federationID,
			PageSize:     defaultListSize,
			PageToken:    pageToken,
		}

		resp, err := config.sdk.OrganizationManager().GroupMapping().ListItems(ctx, req)

		if err != nil {
			return nil, fmt.Errorf("error retrieving group-mapping items for federation %s: %s", federationID, err)
		}

		items = append(items, resp.GroupMappingItems...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return items, nil
}

func groupMappingItemImportStateContext(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) == 3 {
		if err := d.Set("federation_id", parts[0]); err != nil {
			return nil, fmt.Errorf("error setting federation_id: %s", err)
		}
		if err := d.Set("internal_group_id", parts[1]); err != nil {
			return nil, fmt.Errorf("error setting internal_group_id: %s", err)
		}
		if err := d.Set("external_group_id", parts[2]); err != nil {
			return nil, fmt.Errorf("error setting external_group_id: %s", err)
		}
	} else {
		return nil, fmt.Errorf("invalid group mapping item specifier. Expecting {federation_id}/{internal_group_id}/{external_group_id}")
	}
	return []*schema.ResourceData{d}, nil
}
