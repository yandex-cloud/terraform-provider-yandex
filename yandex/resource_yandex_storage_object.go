package yandex

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	if _, err := s3conn.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String(d.Get("acl").(string)),
		Body:   body,
	}); err != nil {
		return fmt.Errorf("error putting object in bucket %q: %s", bucket, err)
	}

	d.SetId(key)

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

	return nil
}

func resourceYandexStorageObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	for _, key := range []string{
		"source",
		"content",
		"content_base64",
	} {
		if d.HasChange(key) {
			return resourceYandexStorageObjectCreate(d, meta)
		}
	}

	if d.HasChange("acl") {
		config := meta.(*Config)
		s3Client, err := getS3Client(d, config)
		if err != nil {
			return fmt.Errorf("error getting storage client: %s", err)
		}
		_, err = s3Client.PutObjectAcl(&s3.PutObjectAclInput{
			Bucket: aws.String(d.Get("bucket").(string)),
			Key:    aws.String(d.Get("key").(string)),
			ACL:    aws.String(d.Get("acl").(string)),
		})
		if err != nil {
			return fmt.Errorf("error putting storage object ACL: %s", err)
		}
	}

	return nil
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

	if _, err := s3client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("error deleting storage object %q in bucket %q: %s ", key, bucket, err)
	}

	return nil
}
