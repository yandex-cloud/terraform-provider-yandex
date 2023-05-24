package yandex

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	homedir "github.com/mitchellh/go-homedir"
)

func resourceYandexStorageObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexStorageObjectCreate,
		Read:   resourceYandexStorageObjectRead,
		Update: resourceYandexStorageObjectUpdate,
		Delete: resourceYandexStorageObjectDelete,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"acl": {
				Type:     schema.TypeString,
				Default:  "private",
				Optional: true,
			},

			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"source": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"content", "content_base64"},
			},

			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content_base64"},
			},

			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"source", "content"},
			},

			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"object_lock_legal_hold_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      s3.ObjectLockLegalHoldStatusOff,
				ValidateFunc: validation.StringInSlice(s3.ObjectLockLegalHoldStatus_Values(), false),
			},

			"object_lock_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nil,
				RequiredWith: []string{"object_lock_retain_until_date"},
				ValidateFunc: validation.StringInSlice(s3.ObjectLockRetentionMode_Values(), false),
			},

			"object_lock_retain_until_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nil,
				RequiredWith: []string{"object_lock_mode"},
				ValidateFunc: validation.IsRFC3339Time,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceYandexStorageObjectCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3conn, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	var body io.ReadSeeker

	if v, ok := d.GetOk("source"); ok {
		source := v.(string)
		path, err := homedir.Expand(source)
		if err != nil {
			return fmt.Errorf("error expanding homedir in source (%s): %s", source, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening storage bucket object source (%s): %s", path, err)
		}

		body = file
		defer func() {
			err := file.Close()
			if err != nil {
				log.Printf("[WARN] Error closing storage bucket object source (%s): %s", path, err)
			}
		}()
	} else if v, ok := d.GetOk("content"); ok {
		content := v.(string)
		body = bytes.NewReader([]byte(content))
	} else if v, ok := d.GetOk("content_base64"); ok {
		content := v.(string)
		// We can't do streaming decoding here (with base64.NewDecoder) because
		// the AWS SDK requires an io.ReadSeeker but a base64 decoder can't seek.
		contentRaw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return fmt.Errorf("error decoding content_base64: %s", err)
		}
		body = bytes.NewReader(contentRaw)
	} else {
		return fmt.Errorf("\"source\", \"content\", or \"content_base64\" field must be specified")
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	log.Printf("[DEBUG] Trying to create new storage object %q in bucket %q", key, bucket)

	awsbucket := aws.String(bucket)
	awskey := aws.String(key)
	putObjectInput := &s3.PutObjectInput{
		Bucket: awsbucket,
		Key:    awskey,
		ACL:    aws.String(d.Get("acl").(string)),
		Body:   body,
	}

	if v, ok := d.GetOk("content_type"); ok {
		putObjectInput.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("object_lock_legal_hold_status"); ok {
		status := v.(string)
		putObjectInput.SetObjectLockLegalHoldStatus(status)
	}

	if v, ok := d.GetOk("object_lock_mode"); ok {
		mode := v.(string)
		v = d.Get("object_lock_retain_until_date")
		// ignore error because the schema has validated the string already
		untilDate, _ := time.Parse(time.RFC3339, v.(string))
		putObjectInput.SetObjectLockMode(mode)
		putObjectInput.SetObjectLockRetainUntilDate(untilDate)
	}

	log.Printf("[DEBUG] Sending putObjectInput %s", putObjectInput.String())

	if _, err := s3conn.PutObject(putObjectInput); err != nil {
		return fmt.Errorf("error putting object in bucket %q: %w", bucket, err)
	}

	d.SetId(key)

	// Use separate request to set tags since it allows to caught
	// NotImplemented error.
	if v, ok := d.GetOk("tags"); ok {
		log.Println("[DEBUG] Trying to set tags for object")

		tags := convertTypesMap(v)
		s3tags := storageBucketTaggingFromMap(tags)

		var req s3.PutObjectTaggingInput
		req.Bucket = awsbucket
		req.Key = awskey
		req.Tagging = &s3.Tagging{
			TagSet: s3tags,
		}

		if _, err = s3conn.PutObjectTagging(&req); err != nil {
			log.Printf("[ERROR] Unable to put S3 Storage Object tags: %v", err)
			return err
		}
	}

	return resourceYandexStorageObjectRead(d, meta)
}

func resourceYandexStorageObjectRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3conn, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	resp, err := s3conn.HeadObject(
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		// If S3 returns a 404 Request Failure, mark the object as destroyed
		if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 404 {
			d.SetId("")
			log.Printf("[WARN] Error Reading Object (%s), object not found (HTTP status 404)", key)
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Reading storage object meta: %s", resp)

	d.Set("content_type", resp.ContentType)

	if resp.ObjectLockLegalHoldStatus != nil {
		status := aws.StringValue(resp.ObjectLockLegalHoldStatus)
		d.Set("object_lock_legal_hold_status", status)
	}

	if resp.ObjectLockMode != nil {
		mode := aws.StringValue(resp.ObjectLockMode)
		untilDate := aws.TimeValue(resp.ObjectLockRetainUntilDate)

		d.Set("object_lock_mode", mode)
		d.Set("object_lock_retain_until_date", untilDate.Format(time.RFC3339))
	}

	tagsResponseRaw, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3conn.GetObjectTagging(&s3.GetObjectTaggingInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: resp.VersionId,
		})
	})
	if err != nil {
		log.Printf("[ERROR] Unable to get S3 Storage Object Tagging: %s", err)
		return err
	}

	tagsResponse := tagsResponseRaw.(*s3.GetObjectTaggingOutput)

	tags := storageBucketTaggingNormalize(tagsResponse.TagSet)
	err = d.Set("tags", tags)
	if err != nil {
		return fmt.Errorf("error setting S3 Storage Object Tagging: %w", err)
	}

	return nil
}

func resourceYandexStorageObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	for _, key := range []string{
		"source",
		"content",
		"content_base64",
		"content_type",
	} {
		if d.HasChange(key) {
			return resourceYandexStorageObjectCreate(d, meta)
		}
	}

	changeHandlers := map[string]func(*s3.S3, *schema.ResourceData) error{
		"acl":                           resourceYandexStorageObjectACLUpdate,
		"object_lock_legal_hold_status": resourceYandexStorageObjectLegalHoldUpdate,
		"object_lock_mode":              resourceYandexStorageObjectRetentionUpdate,
		"object_lock_retain_until_date": resourceYandexStorageObjectRetentionUpdate,
		"tags":                          resourceYandexStorageObjectTaggingUpdate,
	}

	config := meta.(*Config)
	s3Client, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	for name, handler := range changeHandlers {
		if !d.HasChange(name) {
			continue
		}

		err := handler(s3Client, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceYandexStorageObjectACLUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	_, err := s3conn.PutObjectAcl(&s3.PutObjectAclInput{
		Bucket: aws.String(d.Get("bucket").(string)),
		Key:    aws.String(d.Get("key").(string)),
		ACL:    aws.String(d.Get("acl").(string)),
	})
	if err != nil {
		return fmt.Errorf("error putting storage object ACL: %s", err)
	}
	return nil
}

func resourceYandexStorageObjectLegalHoldUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	legalHold := &s3.ObjectLockLegalHold{
		Status: aws.String(d.Get("object_lock_legal_hold_status").(string)),
	}

	_, err := s3conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
		Bucket:    aws.String(d.Get("bucket").(string)),
		Key:       aws.String(d.Get("key").(string)),
		LegalHold: legalHold,
	})
	if err != nil {
		return fmt.Errorf("error putting storage object LegalHoldStatus: %s", err)
	}
	return nil
}

func resourceYandexStorageObjectRetentionUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	retention := &s3.ObjectLockRetention{}

	mode := d.Get("object_lock_mode")
	until := d.Get("object_lock_retain_until_date")

	if mode != nil {
		retention.Mode = aws.String(mode.(string))
		untilDate, _ := time.Parse(time.RFC3339, until.(string))
		retention.RetainUntilDate = aws.Time(untilDate)
	}

	_, err := s3conn.PutObjectRetention(&s3.PutObjectRetentionInput{
		Bucket:                    aws.String(d.Get("bucket").(string)),
		Key:                       aws.String(d.Get("key").(string)),
		Retention:                 retention,
		BypassGovernanceRetention: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("error putting storage object Retention: %s", err)
	}
	return nil
}

func resourceYandexStorageObjectTaggingUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	bucket := aws.String(d.Get("bucket").(string))
	key := aws.String(d.Get("key").(string))

	onUpdate := func(tags []*s3.Tag) error {
		log.Printf("[INFO] Updating Storage S3 object tags with %v", tags)

		request := &s3.PutObjectTaggingInput{
			Bucket: bucket,
			Key:    key,
			Tagging: &s3.Tagging{
				TagSet: tags,
			},
		}
		_, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3conn.PutObjectTagging(request)
		})
		if err != nil {
			log.Printf("[ERROR] Unable to update Storage S3 object tags: %s", err)
		}
		return err
	}

	onDelete := func() error {
		log.Printf("[INFO] Deleting Storage S3 object tags")

		request := &s3.DeleteObjectTaggingInput{
			Bucket: bucket,
			Key:    key,
		}
		_, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3conn.DeleteObjectTagging(request)
		})
		if err != nil {
			log.Printf("[ERROR] Unable to delete Storage S3 object tags: %s", err)
		}
		return err
	}

	return resourceYandexStorageHandleTagsUpdate(d, "object", onUpdate, onDelete)
}

func resourceYandexStorageObjectDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3client, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	// We are effectively ignoring any leading '/' in the key name as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimPrefix(key, "/")

	log.Printf("[DEBUG] Storage Delete Object: %s/%s", bucket, key)

	versionOutput, err := s3client.ListObjectVersions(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error getting version id for deliting storage object %q in bucket %s: %s", key, bucket, err)
	}

	_, err = s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: versionOutput.Versions[0].VersionId,
	})
	if err != nil {
		return fmt.Errorf("error deleting storage object %q in bucket %q: %s ", key, bucket, err)
	}

	return nil
}
