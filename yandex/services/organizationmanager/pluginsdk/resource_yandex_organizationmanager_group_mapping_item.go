package pluginsdk

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const yandexOrganizationManagerGroupMappingItemDefaultTimeout = 1 * time.Minute

func ResourceYandexOrganizationManagerGroupMappingItem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerGroupMappingItemCreate,
		ReadContext:   resourceYandexOrganizationManagerGroupMappingItemRead,
		DeleteContext: resourceYandexOrganizationManagerGroupMappingItemDelete,

		Description: "Allows management of [Yandex Cloud Organization Manager Group Mapping Items](https://yandex.cloud/docs/organization/concepts/add-federation#group-mapping)\n\n ~> Group mapping items depends on [group mapping](organizationmanager_group_mapping.html). If you create group mapping via terraform use `depends_on` meta-argument to avoid errors (see example below)\n.",

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
	config := meta.(*providerconfig.Config)

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
	config := meta.(*providerconfig.Config)

	federationID := d.Get("federation_id").(string)

	items, err := GetGroupMappingItems(context, config, federationID)

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
	config := meta.(*providerconfig.Config)

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

func updateGroupMappingItems(ctx context.Context, config *providerconfig.Config, action organizationmanager.GroupMappingItemDelta_Action, federationID, internalID, externalID string) error {
	client := organizationmanagersdk.NewGroupMappingClient(config.SDK)

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
	log.Printf("[DEBUG] group-mapping update-items request: %s", sdkcommon.ProtoDump(req))

	op, err := client.UpdateItems(ctx, req)
	if err != nil {
		return fmt.Errorf("error while requesting API to update group mapping items op %s: %s", action, err)
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to update group mapping items: %s", err)
	}
	return nil
}

func GetGroupMappingItems(ctx context.Context, config *providerconfig.Config, federationID string) ([]*organizationmanager.GroupMappingItem, error) {
	client := organizationmanagersdk.NewGroupMappingClient(config.SDK)

	items := make([]*organizationmanager.GroupMappingItem, 0)

	pageToken := ""
	for {
		req := &organizationmanager.ListGroupMappingItemsRequest{
			FederationId: federationID,
			PageSize:     sdkcommon.DefaultListSize,
			PageToken:    pageToken,
		}

		resp, err := client.ListItems(ctx, req)

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
