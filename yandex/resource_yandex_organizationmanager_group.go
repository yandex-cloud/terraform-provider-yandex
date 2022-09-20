package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexOrganizationManagerGroupDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerGroupCreate,
		ReadContext:   resourceYandexOrganizationManagerGroupRead,
		UpdateContext: resourceYandexOrganizationManagerGroupUpdate,
		DeleteContext: resourceYandexOrganizationManagerGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerGroupDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerGroupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"organization_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexOrganizationManagerGroupCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	organizationID, err := getOrganizationID(d, config)
	if err != nil {
		return diag.Errorf("Error getting organization ID while creating Group: %s", err)
	}

	req := organizationmanager.CreateGroupRequest{
		OrganizationId: organizationID,
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Group().Create(context, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get Group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*organizationmanager.CreateGroupMetadata)
	if !ok {
		return diag.Errorf("could not get Group ID from create operation metadata")
	}

	d.SetId(md.GroupId)

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error while waiting operation to create Group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("Group creation failed: %s", err)
	}

	return resourceYandexOrganizationManagerGroupRead(context, d, meta)
}

func resourceYandexOrganizationManagerGroupRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenGroup(context, d.Id(), d, meta))
}

func flattenGroup(context context.Context, groupID string, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	group, err := config.sdk.OrganizationManager().Group().Get(context,
		&organizationmanager.GetGroupRequest{
			GroupId: groupID,
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Group %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(group.CreatedAt))
	d.Set("name", group.Name)
	d.Set("organization_id", group.OrganizationId)
	return d.Set("description", group.Description)
}

var updateGroupFieldsMap = map[string]string{
	"name":        "name",
	"description": "description",
}

func resourceYandexOrganizationManagerGroupUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	req := &organizationmanager.UpdateGroupRequest{
		GroupId:     d.Id(),
		UpdateMask:  &field_mask.FieldMask{},
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	var updatePath []string
	for field, path := range updateGroupFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	if len(req.UpdateMask.Paths) == 0 {
		return diag.Errorf("No fields were updated for Group %s", d.Id())
	}

	err := makeGroupUpdateRequest(context, req, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexOrganizationManagerGroupRead(context, d, meta)
}

func resourceYandexOrganizationManagerGroupDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Group %q", d.Id())
	req := &organizationmanager.DeleteGroupRequest{
		GroupId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Group().Delete(context, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Group %q", d.Id())))
	}

	err = op.Wait(context)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Group %q", d.Id())
	return nil
}

func makeGroupUpdateRequest(context context.Context, req *organizationmanager.UpdateGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Group().Update(context, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Group %q: %s", d.Id(), err)
	}

	err = op.Wait(context)
	if err != nil {
		return fmt.Errorf("Error updating Group %q: %s", d.Id(), err)
	}

	return nil
}
