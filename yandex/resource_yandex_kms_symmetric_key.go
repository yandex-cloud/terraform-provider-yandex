package yandex

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

const (
	yandexKMSSymmetricKeyDefaultTimeout = 1 * time.Minute
)

func resourceYandexKMSSymmetricKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexKMSSymmetricKeyCreate,
		Read:   resourceYandexKMSSymmetricKeyRead,
		Update: resourceYandexKMSSymmetricKeyUpdate,
		Delete: resourceYandexKMSSymmetricKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexKMSSymmetricKeyDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexKMSSymmetricKeyDefaultTimeout),
			Update: schema.DefaultTimeout(yandexKMSSymmetricKeyDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexKMSSymmetricKeyDefaultTimeout),
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

			"default_algorithm": {
				Type:         schema.TypeString,
				Default:      "AES_128",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKmsDefaultAlgorithm),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},

			"rotation_period": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateParsableValue(parsePositiveDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rotated_at": {
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
func resourceYandexKMSSymmetricKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating KMS symmetric key: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating KMS symmetric key: %s", err)
	}

	defaultAlgorithm, err := parseKmsDefaultAlgorithm(d.Get("default_algorithm").(string))
	if err != nil {
		return err
	}

	rotationPeriod, err := parseDuration(d.Get("rotation_period").(string))
	if err != nil {
		return err
	}

	req := &kms.CreateSymmetricKeyRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		DefaultAlgorithm:   defaultAlgorithm,
		RotationPeriod:     rotationPeriod,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.KMS().SymmetricKey().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create KMS symmetric key: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get KMS symmetric key create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*kms.CreateSymmetricKeyMetadata)
	if !ok {
		return fmt.Errorf("could not get KMS symmetric key ID from create operation metadata")
	}

	d.SetId(md.KeyId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create KMS symmetric key: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("KMS symmetric key creation failed: %s", err)
	}

	return resourceYandexKMSSymmetricKeyRead(d, meta)
}

func resourceYandexKMSSymmetricKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	key, err := config.sdk.KMS().SymmetricKey().Get(ctx, &kms.GetSymmetricKeyRequest{
		KeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMS Symmetric Key %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(key.CreatedAt))
	d.Set("rotated_at", getTimestamp(key.RotatedAt))
	d.Set("folder_id", key.FolderId)
	d.Set("name", key.Name)
	d.Set("description", key.Description)
	d.Set("default_algorithm", kms.SymmetricAlgorithm_name[int32(key.DefaultAlgorithm)])
	d.Set("rotation_period", formatDuration(key.GetRotationPeriod()))
	d.Set("status", strings.ToLower(key.Status.String()))
	d.Set("deletion_protection", key.DeletionProtection)

	if err := d.Set("labels", key.Labels); err != nil {
		return err
	}

	//TODO support key.PrimaryVersion
	return nil
}

func resourceYandexKMSSymmetricKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var err error
	req := &kms.UpdateSymmetricKeyRequest{
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

	defAlgoName := "default_algorithm"
	if d.HasChange(defAlgoName) {
		defaultAlgorithm, err := parseKmsDefaultAlgorithm(d.Get(defAlgoName).(string))
		if err != nil {
			return err
		}
		req.DefaultAlgorithm = defaultAlgorithm
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, defAlgoName)
	}

	rotationPeriodName := "rotation_period"
	if d.HasChange(rotationPeriodName) {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, rotationPeriodName)
		req.RotationPeriod, err = parseDuration(d.Get("rotation_period").(string))
		if err != nil {
			return err
		}
	}

	deletionProtectionName := "deletion_protection"
	if d.HasChange(deletionProtectionName) {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, deletionProtectionName)
		req.DeletionProtection = d.Get(deletionProtectionName).(bool)
	}

	//TODO support update Status
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.KMS().SymmetricKey().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update KMS Symmetric Key %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating KMS Symmetric Key %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexKMSSymmetricKeyRead(d, meta)
}

func resourceYandexKMSSymmetricKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := config.sdk.KMS().SymmetricKey().Delete(ctx, &kms.DeleteSymmetricKeyRequest{
		KeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMS Symmetric Key %q", d.Id()))
	}

	d.SetId("")
	return nil
}

func parsePositiveDuration(s string) (*duration.Duration, error) {
	d, err := parseDuration(s)
	if err != nil {
		return nil, err
	}

	if d.GetSeconds() == 0 && d.GetNanos() == 0 {
		return nil, fmt.Errorf("can not use zero duration")
	}

	return d, nil
}
