package yandex

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
)

const (
	yandexKMSAsymmetricEncryptionKeyDefaultTimeout = 1 * time.Minute
)

func resourceYandexKMSAsymmetricEncryptionKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexKMSAsymmetricEncryptionKeyCreate,
		Read:   resourceYandexKMSAsymmetricEncryptionKeyRead,
		Update: resourceYandexKMSAsymmetricEncryptionKeyUpdate,
		Delete: resourceYandexKMSAsymmetricEncryptionKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexKMSAsymmetricEncryptionKeyDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexKMSAsymmetricEncryptionKeyDefaultTimeout),
			Update: schema.DefaultTimeout(yandexKMSAsymmetricEncryptionKeyDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexKMSAsymmetricEncryptionKeyDefaultTimeout),
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"encryption_algorithm": {
				Type:         schema.TypeString,
				Default:      "RSA_2048_ENC_OAEP_SHA_256",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKmsAsymmetricEncryptionAlgorithm),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceYandexKMSAsymmetricEncryptionKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating KMS asymmetric encryption key: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating KMS asymmetric encryption key: %s", err)
	}

	encryptionAlgorithm, err := parseKmsAsymmetricEncryptionAlgorithm(d.Get("encryption_algorithm").(string))
	if err != nil {
		return err
	}

	req := &kms.CreateAsymmetricEncryptionKeyRequest{
		FolderId:            folderID,
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Labels:              labels,
		EncryptionAlgorithm: encryptionAlgorithm,
		DeletionProtection:  d.Get("deletion_protection").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create KMS asymmetric encryption key: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get KMS asymmetric encryption key create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*kms.CreateAsymmetricEncryptionKeyMetadata)
	if !ok {
		return fmt.Errorf("could not get KMS asymmetric encryption key ID from create operation metadata")
	}

	d.SetId(md.KeyId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create KMS asymmetric encryption key: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("KMS asymmetric encryption key creation failed: %s", err)
	}

	return resourceYandexKMSAsymmetricEncryptionKeyRead(d, meta)
}

func resourceYandexKMSAsymmetricEncryptionKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	key, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Get(ctx, &kms.GetAsymmetricEncryptionKeyRequest{
		KeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMS AsymmetricEncryptionKey %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(key.CreatedAt))
	d.Set("folder_id", key.FolderId)
	d.Set("name", key.Name)
	d.Set("description", key.Description)
	d.Set("encryption_algorithm", kms.AsymmetricEncryptionAlgorithm_name[int32(key.EncryptionAlgorithm)])
	d.Set("status", strings.ToLower(key.Status.String()))
	d.Set("deletion_protection", key.DeletionProtection)

	if err := d.Set("labels", key.Labels); err != nil {
		return err
	}

	return nil
}

func resourceYandexKMSAsymmetricEncryptionKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var err error
	req := &kms.UpdateAsymmetricEncryptionKeyRequest{
		KeyId:      d.Id(),
		UpdateMask: &field_mask.FieldMask{},
	}

	d.Partial(true)

	labelPropName := "labels"
	if d.HasChange(labelPropName) {
		labelsProp, err := expandLabels(d.Get(labelPropName))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, labelPropName)
	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req.Name = d.Get(namePropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, namePropName)
	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req.Description = d.Get(descPropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, descPropName)
	}

	deletionProtectionName := "deletion_protection"
	if d.HasChange(deletionProtectionName) {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, deletionProtectionName)
		req.DeletionProtection = d.Get(deletionProtectionName).(bool)
	}

	//TODO support update Status
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update KMS AsymmetricEncryptionKey %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating KMS AsymmetricEncryptionKey %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexKMSAsymmetricEncryptionKeyRead(d, meta)
}

func resourceYandexKMSAsymmetricEncryptionKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Delete(ctx, &kms.DeleteAsymmetricEncryptionKeyRequest{
		KeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMS AsymmetricEncryptionKey %q", d.Id()))
	}

	d.SetId("")
	return nil
}
