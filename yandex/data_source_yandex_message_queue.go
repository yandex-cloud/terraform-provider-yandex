package yandex

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func dataSourceYandexMessageQueue() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMessageQueueRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// Credentials
			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexMessageQueueRead(d *schema.ResourceData, meta interface{}) error {
	ymqClient, err := newYMQClient(d, meta)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)

	log.Printf("[INFO] Getting queue url of queue %s", name)

	var urlOutput *sqs.GetQueueUrlOutput
	// TODO: SA1019: resource.Retry is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
	err = resource.Retry(15*time.Second, func() *resource.RetryError {
		urlOutput, err = ymqClient.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(name),
		})

		if err != nil {
			// Queue can be not found immediately after its creation.
			// It can occur in not found or access denied exception.
			if isAWSSQSErr(err, sqs.ErrCodeQueueDoesNotExist) || isAWSSQSErr(err, "AccessDeniedException") {
				// TODO: SA1019: resource.RetryableError is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
				return resource.RetryableError(err)
			}
			// TODO: SA1019: resource.NonRetryableError is deprecated: Use helper/retry package instead. This is required for migrating acceptance testing to terraform-plugin-testing. (staticcheck)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil || urlOutput.QueueUrl == nil {
		return fmt.Errorf("Error getting queue url: %s", err)
	}

	queueURL := aws.StringValue(urlOutput.QueueUrl)

	var attributesOutput *sqs.GetQueueAttributesOutput
	err = resource.Retry(15*time.Second, func() *resource.RetryError {
		attributesOutput, err = ymqClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(queueURL),
			AttributeNames: []*string{aws.String(sqs.QueueAttributeNameQueueArn)},
		})

		if err != nil {
			// Queue can be not found immediately after its creation.
			// It can occur in not found or access denied exception.
			if isAWSSQSErr(err, sqs.ErrCodeQueueDoesNotExist) || isAWSSQSErr(err, "AccessDeniedException") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error getting queue attributes: %s", err)
	}

	d.Set("arn", aws.StringValue(attributesOutput.Attributes[sqs.QueueAttributeNameQueueArn]))
	d.Set("url", queueURL)
	d.SetId(queueURL)

	return nil
}
