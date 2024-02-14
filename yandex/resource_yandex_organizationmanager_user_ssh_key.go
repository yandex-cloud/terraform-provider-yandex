package yandex

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"

	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexOrganizationManagerUserSshKeyDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerUserSshKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexOrganizationManagerUserSshKeyCreate,
		ReadContext:   resourceYandexOrganizationManagerUserSshKeyRead,
		UpdateContext: resourceYandexOrganizationManagerUserSshKeyUpdate,
		DeleteContext: resourceYandexOrganizationManagerUserSshKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerUserSshKeyDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerUserSshKeyDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerUserSshKeyDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerUserSshKeyDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"subject_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"data": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 20000),
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"organization_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceYandexOrganizationManagerUserSshKeyRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenUserSshKey(context, d, meta))
}

func resourceYandexOrganizationManagerUserSshKeyCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	organizationID := d.Get("organization_id").(string)
	subjectID := d.Get("subject_id").(string)
	log.Printf("[INFO] Creating user ssh key for organization %q and subject %q", organizationID, subjectID)

	req := &organizationmanager.CreateUserSshKeyRequest{
		OrganizationId: organizationID,
		SubjectId:      subjectID,
		Data:           d.Get("data").(string),
	}

	if v, ok := d.GetOk("name"); ok {
		req.SetName(v.(string))
	}

	if v, ok := d.GetOk("expires_at"); ok {
		expiresAt, err := parseTimestamp(v.(string))
		if err != nil {
			return diag.Errorf("Error during parsing field expires_at while creating user ssh key for organization %q and subject %q", organizationID, subjectID)
		}
		req.SetExpiresAt(expiresAt)
	}

	config := meta.(*Config)
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().UserSshKey().Create(context, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create user ssh key for organization %q and subject %q: %s", organizationID, subjectID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error creating user ssh key for organization %q and subject %q: %s", organizationID, subjectID, err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get user ssh key create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*organizationmanager.CreateUserSshKeyMetadata)
	if !ok {
		return diag.Errorf("Could not get user ssh key ID from create operation metadata")
	}

	d.SetId(md.UserSshKeyId)

	log.Printf("[INFO] User ssh key %q was created", md.UserSshKeyId)

	return resourceYandexOrganizationManagerUserSshKeyRead(context, d, meta)
}

var updateUserSshKeyFieldsMap = map[string]string{
	"name":       "name",
	"expires_at": "expires_at",
}

func resourceYandexOrganizationManagerUserSshKeyUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userSSHKeyID := d.Id()
	res, ok := d.GetOk("user_ssh_key_id")
	if ok {
		userSSHKeyID = res.(string)
	}

	log.Printf("[INFO] Updating user ssh key %q", userSSHKeyID)

	req := &organizationmanager.UpdateUserSshKeyRequest{
		UserSshKeyId: userSSHKeyID,
		UpdateMask:   &field_mask.FieldMask{},
	}

	if d.HasChange("name") {
		req.SetName(d.Get("name").(string))
	}

	if d.HasChange("expires_at") {
		expiresAt, err := parseTimestamp(d.Get("expires_at").(string))
		if err != nil {
			return diag.Errorf("Error during parsing field expires_at for user ssh key %q", userSSHKeyID)
		}
		req.SetExpiresAt(expiresAt)
	}

	var updatePath []string
	for field, path := range updateUserSshKeyFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	if len(req.UpdateMask.Paths) == 0 {
		return diag.Errorf("No fields were updated for  user ssh key %q", userSSHKeyID)
	}

	config := meta.(*Config)
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().UserSshKey().Update(context, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to update user ssh key %q: %s", userSSHKeyID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error updating user ssh key %q: %s", userSSHKeyID, err)
	}

	d.SetId(userSSHKeyID)
	log.Printf("[INFO] User ssh key %q was updated", userSSHKeyID)

	return resourceYandexOrganizationManagerUserSshKeyRead(context, d, meta)
}

func resourceYandexOrganizationManagerUserSshKeyDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userSSHKeyID := d.Id()
	res, ok := d.GetOk("user_ssh_key_id")
	if ok {
		userSSHKeyID = res.(string)
	}

	log.Printf("[INFO] Deleting user ssh key %q", userSSHKeyID)

	req := &organizationmanager.DeleteUserSshKeyRequest{
		UserSshKeyId: userSSHKeyID,
	}

	config := meta.(*Config)
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().UserSshKey().Delete(context, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to delete user ssh key %q: %s", userSSHKeyID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error deleting user ssh key %q: %s", userSSHKeyID, err)
	}

	log.Printf("[INFO] User ssh key %q was deleted", userSSHKeyID)

	return nil
}
