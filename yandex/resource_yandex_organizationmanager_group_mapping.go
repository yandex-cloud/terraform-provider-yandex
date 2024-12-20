package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexOrganizationManagerGroupMappingDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerGroupMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerGroupMappingCreate,
		ReadContext:   resourceYandexOrganizationManagerGroupMappingRead,
		UpdateContext: resourceYandexOrganizationManagerGroupMappingUpdate,
		DeleteContext: resourceYandexOrganizationManagerGroupMappingDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerGroupMappingDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerGroupMappingDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerGroupMappingDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerGroupMappingDefaultTimeout),
		},

		Importer: &schema.ResourceImporter{
			StateContext: groupMappingImportStateContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set \"true\" to enable organization manager group mapping",
			},

			"federation_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
				Description:  "ID of the SAML Federation",
			},
		},
	}
}

func resourceYandexOrganizationManagerGroupMappingCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	federationID := d.Get("federation_id").(string)

	req := &organizationmanager.CreateGroupMappingRequest{
		FederationId: federationID,
		Enabled:      d.Get("enabled").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().GroupMapping().Create(context, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create group mapping: %s", err)
	}

	d.SetId(federationID)

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error while waiting operation to create group mapping: %s", err)
	}

	return resourceYandexOrganizationManagerGroupMappingRead(context, d, meta)
}

func resourceYandexOrganizationManagerGroupMappingRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &organizationmanager.GetGroupMappingRequest{
		FederationId: d.Get("federation_id").(string),
	}

	resp, err := config.sdk.OrganizationManager().GroupMapping().Get(context, req)

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("group_mapping %q", d.Id())))
	}

	d.Set("federation_id", resp.GetGroupMapping().GetFederationId())
	d.Set("enabled", resp.GetGroupMapping().GetEnabled())

	return nil
}

func resourceYandexOrganizationManagerGroupMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &organizationmanager.UpdateGroupMappingRequest{
		UpdateMask:   &field_mask.FieldMask{},
		FederationId: d.Get("federation_id").(string),
		Enabled:      d.Get("enabled").(bool),
	}

	if d.HasChange("enabled") {
		req.Enabled = d.Get("enabled").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "enabled")
	}

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().GroupMapping().Update(ctx, req))
		if err != nil {
			return diag.Errorf("Error while requesting API to update group mapping: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return diag.Errorf("Error while waiting operation to update group mapping: %s", err)

		}

		if _, err := op.Response(); err != nil {
			return diag.Errorf("Group mapping update failed: %s", err)
		}
	}

	return resourceYandexOrganizationManagerGroupMappingRead(ctx, d, meta)
}

func resourceYandexOrganizationManagerGroupMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &organizationmanager.DeleteGroupMappingRequest{
		FederationId: d.Get("federation_id").(string),
	}

	log.Printf("[DEBUG] Delete GroupMapping request: %s", protoDump(req))

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().GroupMapping().Delete(ctx, req))

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("group_mapping %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func groupMappingImportStateContext(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("federation_id", d.Id()); err != nil {
		return nil, fmt.Errorf("error setting federation_id: %s", err)
	}
	return []*schema.ResourceData{d}, nil
}
