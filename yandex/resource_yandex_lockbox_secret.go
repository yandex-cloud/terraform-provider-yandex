package yandex

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	yandexLockboxSecretDefaultTimeout = 1 * time.Minute
)

func resourceYandexLockboxSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexLockboxSecretCreate,
		ReadContext:   resourceYandexLockboxSecretRead,
		UpdateContext: resourceYandexLockboxSecretUpdate,
		DeleteContext: resourceYandexLockboxSecretDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLockboxSecretDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexLockboxSecretDefaultTimeout),
			Update: schema.DefaultTimeout(yandexLockboxSecretDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLockboxSecretDefaultTimeout),
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"folder_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile("^([-_0-9a-z]*)$"), ""), validation.StringLenBetween(0, 63)),
				},
				Set:      schema.HashString,
				Optional: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"password_payload_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password_key": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexp.MustCompile(`^[-_./\\@0-9a-zA-Z]+$`), ""),
							),
						},
						"length": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      36,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"include_uppercase": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_lowercase": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_digits": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_punctuation": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"included_punctuation": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"password_payload_specification.0.excluded_punctuation"},
							ValidateFunc:  validation.StringLenBetween(0, 32),
						},
						"excluded_punctuation": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"password_payload_specification.0.included_punctuation"},
							ValidateFunc:  validation.StringLenBetween(0, 31),
						},
					},
				},
			},
		},
	}
}

func resourceYandexLockboxSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	var payloadSpecification lockbox.CreateSecretRequest_PayloadSpecification = nil

	pps, err := expandPasswordPayloadSpecification(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if pps != nil {
		payloadSpecification = &lockbox.CreateSecretRequest_PasswordPayloadSpecification{
			PasswordPayloadSpecification: pps,
		}
	}

	req := &lockbox.CreateSecretRequest{
		FolderId:             folderID,
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Labels:               expandStringStringMap(d.Get("labels").(map[string]interface{})),
		KmsKeyId:             d.Get("kms_key_id").(string),
		DeletionProtection:   d.Get("deletion_protection").(bool),
		PayloadSpecification: payloadSpecification,
		CreateVersion:        &wrapperspb.BoolValue{Value: false},
	}

	log.Printf("[INFO] creating Lockbox secret: %s", protojson.Format(req))

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().Create(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to create secret: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of create secret: %s", err)
	}

	md, ok := protoMetadata.(*lockbox.CreateSecretMetadata)
	if !ok {
		return diag.Errorf("could not get Secret ID from create operation metadata")
	}

	d.SetId(md.SecretId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to create secret: %s", err)

	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("secret creation failed: %s", err)
	}

	log.Printf("[INFO] created Lockbox secret with ID: %s", d.Id())

	return resourceYandexLockboxSecretRead(ctx, d, meta)
}

func resourceYandexLockboxSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return yandexLockboxSecretRead(d.Id(), false, ctx, d, meta)
}

// read is almost the same in resource and data source
func yandexLockboxSecretRead(id string, isDataSource bool, ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &lockbox.GetSecretRequest{
		SecretId: id,
	}

	log.Printf("[INFO] reading Lockbox secret: %s", protojson.Format(req))

	secret, err := config.sdk.LockboxSecret().Secret().Get(ctx, req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("secret %q", id)))
	}

	// Specific logic for data source
	if isDataSource {
		d.SetId(req.SecretId)

		currentVersion, err := flattenLockboxVersion(secret.GetCurrentVersion())
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("current_version", currentVersion); err != nil {
			log.Printf("[ERROR] failed set field current_version: %s", err)
			return diag.FromErr(err)
		}
	}

	createdAt := getTimestamp(secret.GetCreatedAt())

	if err := d.Set("created_at", createdAt); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("deletion_protection", secret.GetDeletionProtection()); err != nil {
		log.Printf("[ERROR] failed set field deletion_protection: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("description", secret.GetDescription()); err != nil {
		log.Printf("[ERROR] failed set field description: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", secret.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("kms_key_id", secret.GetKmsKeyId()); err != nil {
		log.Printf("[ERROR] failed set field kms_key_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("labels", secret.GetLabels()); err != nil {
		log.Printf("[ERROR] failed set field labels: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("name", secret.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("status", secret.GetStatus().String()); err != nil {
		log.Printf("[ERROR] failed set field status: %s", err)
		return diag.FromErr(err)
	}

	passwordPayloadSpecification := flattenPasswordPayloadSpecification(secret.GetPasswordPayloadSpecification())
	if err = d.Set("password_payload_specification", passwordPayloadSpecification); err != nil {
		log.Printf("[ERROR] failed set field password_payload_specification: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] read Lockbox secret with ID: %s", d.Id())

	return nil
}

func resourceYandexLockboxSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	var payloadSpecification lockbox.UpdateSecretRequest_PayloadSpecification = nil

	pps, err := expandPasswordPayloadSpecification(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if pps != nil {
		payloadSpecification = &lockbox.UpdateSecretRequest_PasswordPayloadSpecification{
			PasswordPayloadSpecification: pps,
		}
	}

	req := &lockbox.UpdateSecretRequest{
		SecretId:             d.Id(),
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Labels:               expandStringStringMap(d.Get("labels").(map[string]interface{})),
		DeletionProtection:   d.Get("deletion_protection").(bool),
		PayloadSpecification: payloadSpecification,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: generateFieldMasks(d, resourceYandexLockboxSecretUpdateFieldsMap),
		},
	}

	log.Printf("[INFO] updating Lockbox secret: %s", protojson.Format(req))

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().Update(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to update secret: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to update secret: %s", err)

	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("secret update failed: %s", err)
	}

	log.Printf("[INFO] updated Lockbox secret with ID: %s", d.Id())

	return resourceYandexLockboxSecretRead(ctx, d, meta)
}

func resourceYandexLockboxSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &lockbox.DeleteSecretRequest{
		SecretId: d.Id(),
	}

	log.Printf("[INFO] deleting Lockbox secret: %s", protojson.Format(req))

	op, err := config.sdk.WrapOperation(config.sdk.LockboxSecret().Secret().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("secret %q", d.Id())))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted Lockbox secret with ID: %s", d.Id())

	return nil
}

var resourceYandexLockboxSecretUpdateFieldsMap = map[string]string{
	"name":                           "name",
	"description":                    "description",
	"labels":                         "labels",
	"deletion_protection":            "deletion_protection",
	"password_payload_specification": "password_payload_specification",
}
