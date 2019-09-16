package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

func resourceYandexIAMServiceAccountStaticAccessKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIAMServiceAccountStaticAccessKeyCreate,
		Read:   resourceYandexIAMServiceAccountStaticAccessKeyRead,
		Delete: resourceYandexIAMServiceAccountStaticAccessKeyDelete,

		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// There is no Update method for IAM SA Key resource,
			// so "description" attr set as 'ForceNew:true'
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexIAMServiceAccountStaticAccessKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := config.sdk.IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: d.Get("service_account_id").(string),
		Description:      d.Get("description").(string),
	})

	if err != nil {
		return fmt.Errorf("Error create service account key: %s", err)
	}

	d.SetId(resp.AccessKey.Id)
	// One-time generated value
	d.Set("secret_key", resp.Secret)

	return resourceYandexIAMServiceAccountStaticAccessKeyRead(d, meta)
}

func resourceYandexIAMServiceAccountStaticAccessKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	sak, err := config.sdk.IAM().AWSCompatibility().AccessKey().Get(ctx, &awscompatibility.GetAccessKeyRequest{
		AccessKeyId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Key %q", d.Id()))
	}

	ts, err := ptypes.Timestamp(sak.CreatedAt)
	if err != nil {
		return fmt.Errorf("error while convert CreatedAt timestamp: %s", err)
	}

	d.Set("access_key", sak.KeyId)
	d.Set("description", sak.Description)
	d.Set("service_account_id", sak.ServiceAccountId)
	d.Set("created_at", ts.Format(time.RFC3339))

	return nil
}

func resourceYandexIAMServiceAccountStaticAccessKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.ContextWithClientTraceID(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Delete(ctx, &awscompatibility.DeleteAccessKeyRequest{
		AccessKeyId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Key %q", d.Id()))
	}

	d.SetId("")
	return nil
}
