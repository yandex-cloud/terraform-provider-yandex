package yandex

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceYandexStorageObject() *schema.Resource {
	return &schema.Resource{
		Description:   "Allows management of [Yandex Cloud Storage Object](https://yandex.cloud/docs/storage/concepts/object).",
		CreateContext: resourceYandexStorageObjectCreate,
		ReadContext:   resourceYandexStorageObjectRead,
		UpdateContext: resourceYandexStorageObjectUpdate,
		DeleteContext: resourceYandexStorageObjectDelete,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:        schema.TypeString,
				Description: "The name of the containing bucket.",
				Required:    true,
				ForceNew:    true,
			},

			"access_key": {
				Type:        schema.TypeString,
				Description: "The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:    true,
				Sensitive:   true,
			},

			"secret_key": {
				Type:        schema.TypeString,
				Description: "The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:    true,
				Sensitive:   true,
			},

			"acl": {
				Type:        schema.TypeString,
				Description: "The [predefined ACL](https://yandex.cloud/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`.\n\n~> To change ACL after creation, the service account to which used access and secret keys correspond should have `storage.admin` role, though this role is not necessary to be able to create an object with any ACL.\n",
				Default:     "private",
				Optional:    true,
			},

			"key": {
				Type:        schema.TypeString,
				Description: "The name of the object once it is in the bucket.",
				Required:    true,
				ForceNew:    true,
			},

			"source": {
				Type:          schema.TypeString,
				Description:   "The path to a file that will be read and uploaded as raw bytes for the object content. Conflicts with `content` and `content_base64`.",
				Optional:      true,
				ConflictsWith: []string{"content", "content_base64"},
			},

			"source_hash": {
				Type:        schema.TypeString,
				Description: "Used to trigger object update when the source content changes. So the only meaningful value is `filemd5(\"path/to/source\"). The value is only stored in state and not saved by Yandex Storage.",
				Optional:    true,
			},

			"content": {
				Type:          schema.TypeString,
				Description:   "Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text. Conflicts with `source` and `content_base64`.",
				Optional:      true,
				ConflictsWith: []string{"source", "content_base64"},
			},

			"content_base64": {
				Type:          schema.TypeString,
				Description:   "Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file. Conflicts with `source` and `content`.",
				Optional:      true,
				ConflictsWith: []string{"source", "content"},
			},

			"content_type": {
				Type:        schema.TypeString,
				Description: "A standard MIME type describing the format of the object data, e.g. `application/octet-stream`. All Valid MIME Types are valid for this input.",
				Optional:    true,
				Computed:    true,
			},

			"object_lock_legal_hold_status": {
				Type:         schema.TypeString,
				Description:  "Specifies a [legal hold status](https://yandex.cloud/docs/storage/concepts/object-lock#types) of an object. Requires `object_lock_configuration` to be enabled on a bucket.",
				Optional:     true,
				Default:      nil,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockLegalHoldStatusValues, false),
			},

			"object_lock_mode": {
				Type:         schema.TypeString,
				Description:  "Specifies a type of object lock. One of `[\"GOVERNANCE\", \"COMPLIANCE\"]`. It must be set simultaneously with `object_lock_retain_until_date`. Requires `object_lock_configuration` to be enabled on a bucket.",
				Optional:     true,
				Default:      nil,
				RequiredWith: []string{"object_lock_retain_until_date"},
				ValidateFunc: validation.StringInSlice(s3.ObjectLockRetentionModeValues, false),
			},

			"object_lock_retain_until_date": {
				Type:         schema.TypeString,
				Description:  "Specifies date and time in RTC3339 format until which an object is to be locked. It must be set simultaneously with `object_lock_mode`. Requires `object_lock_configuration` to be enabled on a bucket.",
				Optional:     true,
				Default:      nil,
				RequiredWith: []string{"object_lock_mode"},
				ValidateFunc: validation.IsRFC3339Time,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceYandexStorageObjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return diag.Errorf("error getting storage client: %s", err)
	}

	data := s3.CreationData{
		Bucket: d.Get("bucket").(string),
		Key:    d.Get("key").(string),
		ACL:    d.Get("acl").(string),
	}

	if v, ok := d.GetOk("source"); ok {
		data.Source = &s3.Source{
			Type:  s3.SourceTypeFile,
			Value: v.(string),
		}
	} else if v, ok := d.GetOk("content"); ok {
		data.Source = &s3.Source{
			Type:  s3.SourceTypeContent,
			Value: v.(string),
		}
	} else if v, ok := d.GetOk("content_base64"); ok {
		data.Source = &s3.Source{
			Type:  s3.SourceTypeContentBase64,
			Value: v.(string),
		}
	} else {
		return diag.Errorf("\"source\", \"content\", or \"content_base64\" field must be specified")
	}

	if v, ok := d.GetOk("content_type"); ok {
		data.ContentType = v.(string)
	}
	if v, ok := d.GetOk("object_lock_legal_hold_status"); ok {
		data.ObjectLockLegalHoldStatus = v.(string)
	}
	if v, ok := d.GetOk("object_lock_mode"); ok {
		mode := v.(string)
		untilDate, err := time.Parse(time.RFC3339, d.Get("object_lock_retain_until_date").(string))
		if err != nil {
			return diag.Errorf("error parsing object_lock_retain_until_date: %s", err)
		}
		data.ObjectRetention = &s3.ObjectRetention{
			Mode:            mode,
			RetainUntilDate: untilDate,
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		data.Tags = s3.NewTags(v)
	}

	log.Printf("[DEBUG] Trying to create new storage object %q in bucket %q", data.Key, data.Bucket)
	isCreated, err := s3Client.CreateObject(ctx, data)
	if isCreated {
		d.SetId(data.Key)
	}
	if err != nil {
		log.Printf("[ERROR] Unable to create S3 Storage Object: %v", err)
		return diag.Errorf("error creating storage object: %s", err)
	}

	return resourceYandexStorageObjectRead(ctx, d, meta)
}

func resourceYandexStorageObjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return diag.Errorf("error getting storage client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	object, err := s3Client.GetObject(ctx, bucket, key)
	if err != nil {
		// If S3 returns a 404 Request Failure, mark the object as destroyed
		if errors.Is(err, s3.ErrObjectNotFound) {
			log.Printf("[WARN] Error Reading Object (%s), object not found (HTTP status 404)", key)
			d.SetId("")
			return nil
		}
		log.Printf("[ERROR] Unable to get S3 Storage Object: %s", err)
		return diag.FromErr(err)
	}

	d.Set("content_type", object.ContentType)
	if object.ObjectLockLegalHoldStatus != nil {
		d.Set("object_lock_legal_hold_status", *object.ObjectLockLegalHoldStatus)
	}
	if object.ObjectRetention != nil {
		d.Set("object_lock_mode", object.ObjectRetention.Mode)
		d.Set("object_lock_retain_until_date", object.ObjectRetention.RetainUntilDate.Format(time.RFC3339))
	}
	err = d.Set("tags", s3.TagsToRaw(object.Tags))
	if err != nil {
		return diag.Errorf("error setting S3 Storage Object Tagging: %s", err)
	}

	return nil
}

func resourceYandexStorageObjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if hasObjectContentChanged(d) {
		return resourceYandexStorageObjectCreate(ctx, d, meta)
	}

	changeHandlers := map[string]func(context.Context, *s3.Client, *schema.ResourceData) error{
		"acl":                           resourceYandexStorageObjectACLUpdate,
		"object_lock_legal_hold_status": resourceYandexStorageObjectLegalHoldUpdate,
		"tags":                          resourceYandexStorageObjectTaggingUpdate,
	}

	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return diag.Errorf("error getting storage client: %s", err)
	}

	for name, handler := range changeHandlers {
		if !d.HasChange(name) {
			continue
		}

		err := handler(ctx, s3Client, d)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges("object_lock_mode", "object_lock_retain_until_date") {
		if err := resourceYandexStorageObjectRetentionUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func hasObjectContentChanged(d *schema.ResourceData) bool {
	for _, key := range []string{
		"source",
		"source_hash",
		"content",
		"content_base64",
		"content_type",
	} {
		if d.HasChange(key) {
			return true
		}
	}

	return false
}

func resourceYandexStorageObjectACLUpdate(ctx context.Context, s3Client *s3.Client, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	acl := d.Get("acl").(string)

	return s3Client.UpdateObjectACL(ctx, bucket, key, acl)
}

func resourceYandexStorageObjectLegalHoldUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	status := d.Get("object_lock_legal_hold_status").(string)

	return s3Client.UpdateObjectLegalHold(ctx, bucket, key, status)
}

func resourceYandexStorageObjectRetentionUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	mode := d.Get("object_lock_mode")

	var retention *s3.ObjectRetention
	if mode != nil {
		untilDate, _ := time.Parse(time.RFC3339, d.Get("object_lock_retain_until_date").(string))
		retention = &s3.ObjectRetention{
			Mode:            mode.(string),
			RetainUntilDate: untilDate,
		}
	}

	return s3Client.UpdateObjectRetention(ctx, bucket, key, retention)
}

func resourceYandexStorageObjectTaggingUpdate(ctx context.Context, s3Client *s3.Client, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	tags := s3.NewTags(d.Get("tags"))
	if err := s3Client.UpdateObjectTags(ctx, bucket, key, tags); err != nil {
		log.Printf("[ERROR] Unable to update Storage S3 object: %s", err)
		return err
	}
	return nil
}

func resourceYandexStorageObjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return diag.Errorf("error getting storage client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	// We are effectively ignoring any leading '/' in the key name as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimPrefix(key, "/")

	log.Printf("[DEBUG] Storage Delete Object: %s/%s", bucket, key)
	if err := s3Client.DeleteObject(ctx, bucket, key); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
