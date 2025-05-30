package yandex

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const defaultYMQRegion = "ru-central1"

var sqsQueueAttributeMap = map[string]string{
	"delay_seconds":               sqs.QueueAttributeNameDelaySeconds,
	"max_message_size":            sqs.QueueAttributeNameMaximumMessageSize,
	"message_retention_seconds":   sqs.QueueAttributeNameMessageRetentionPeriod,
	"receive_wait_time_seconds":   sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds,
	"visibility_timeout_seconds":  sqs.QueueAttributeNameVisibilityTimeout,
	"redrive_policy":              sqs.QueueAttributeNameRedrivePolicy,
	"arn":                         sqs.QueueAttributeNameQueueArn,
	"fifo_queue":                  sqs.QueueAttributeNameFifoQueue,
	"content_based_deduplication": sqs.QueueAttributeNameContentBasedDeduplication,
}

func resourceYandexMessageQueue() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud Message Queue](https://yandex.cloud/docs/message-queue).",

		Create: resourceYandexMessageQueueCreate,
		Read:   resourceYandexMessageQueueRead,
		Update: resourceYandexMessageQueueUpdate,
		Delete: resourceYandexMessageQueueDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Description:   "Queue name. The maximum length is 80 characters. You can use numbers, letters, underscores, and hyphens in the name. The name of a FIFO queue must end with the `.fifo` suffix. If not specified, random name will be generated. Conflicts with `name_prefix`. For more information see [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/CreateQueue).",
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateQueueName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Description:   "Generates random name with the specified prefix. Conflicts with `name`.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"delay_seconds": {
				Type:         schema.TypeInt,
				Description:  "Number of seconds to [delay the message from being available for processing](https://yandex.cloud/docs/message-queue/concepts/delay-queues#delay-queues). Valid values: from 0 to 900 seconds (15 minutes). Default: 0.",
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 900),
			},
			"max_message_size": {
				Type:         schema.TypeInt,
				Description:  "Maximum message size in bytes. Valid values: from 1024 bytes (1 KB) to 262144 bytes (256 KB). Default: 262144 (256 KB). For more information see [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/CreateQueue).",
				Optional:     true,
				Default:      262144,
				ValidateFunc: validation.IntBetween(1024, 262144),
			},
			"message_retention_seconds": {
				Type:         schema.TypeInt,
				Description:  "The length of time in seconds to retain a message. Valid values: from 60 seconds (1 minute) to 1209600 seconds (14 days). Default: 345600 (4 days). For more information see [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/CreateQueue).",
				Optional:     true,
				Default:      345600,
				ValidateFunc: validation.IntBetween(60, 1209600),
			},
			"receive_wait_time_seconds": {
				Type:         schema.TypeInt,
				Description:  "Wait time for the [ReceiveMessage](https://yandex.cloud/docs/message-queue/api-ref/message/ReceiveMessage) method (for long polling), in seconds. Valid values: from 0 to 20 seconds. Default: 0. For more information about long polling see [documentation](https://yandex.cloud/docs/message-queue/concepts/long-polling).",
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 20),
			},
			"visibility_timeout_seconds": {
				Type:         schema.TypeInt,
				Description:  "[Visibility timeout](https://yandex.cloud/docs/message-queue/concepts/visibility-timeout) for messages in a queue, specified in seconds. Valid values: from 0 to 43200 seconds (12 hours). Default: 30.",
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(0, 43200),
			},
			"redrive_policy": {
				Type:         schema.TypeString,
				Description:  "Message redrive policy in [Dead Letter Queue](https://yandex.cloud/docs/message-queue/concepts/dlq). The source queue and DLQ must be the same type: for FIFO queues, the DLQ must also be a FIFO queue. For more information about redrive policy see [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/CreateQueue). Also you can use example in this page.",
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"fifo_queue": {
				Type:        schema.TypeBool,
				Description: "Is this queue [FIFO](https://yandex.cloud/docs/message-queue/concepts/queue#fifo-queues). If this parameter is not used, a standard queue is created. You cannot change the parameter value for a created queue.",
				Default:     false,
				ForceNew:    true,
				Optional:    true,
			},
			"content_based_deduplication": {
				Type:        schema.TypeBool,
				Description: "Enables [content-based deduplication](https://yandex.cloud/docs/message-queue/concepts/deduplication#content-based-deduplication). Can be used only if queue is [FIFO](https://yandex.cloud/docs/message-queue/concepts/queue#fifo-queues).",
				Default:     false,
				Optional:    true,
			},
			"region_id": {
				Type:        schema.TypeString,
				Description: "ID of the region where the message queue is located at. The default is 'ru-central1'.",
				Optional:    true,
				Default:     defaultYMQRegion,
				ForceNew:    true,
			},
			"tags": {
				Type:        schema.TypeMap,
				Description: "SQS tags",
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
				Optional:    true,
			},

			// Credentials
			"access_key": {
				Type:        schema.TypeString,
				Description: "The [access key](https://yandex.cloud/docs/iam/operations/sa/create-access-key) to use when applying changes. If omitted, `ymq_access_key` specified in provider config is used. For more information see [documentation](https://yandex.cloud/docs/message-queue/quickstart).",
				Optional:    true,
			},
			"secret_key": {
				Type:        schema.TypeString,
				Description: "The [secret key](https://yandex.cloud/docs/iam/operations/sa/create-access-key) to use when applying changes. If omitted, `ymq_secret_key` specified in provider config is used. For more information see [documentation](https://yandex.cloud/docs/message-queue/quickstart).",
				Optional:    true,
				Sensitive:   true,
			},

			// Computed
			"arn": {
				Type:        schema.TypeString,
				Description: "ARN of the Yandex Message Queue. It is used for setting up a [redrive policy](https://yandex.cloud/docs/message-queue/concepts/dlq). See [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/SetQueueAttributes).",
				Computed:    true,
			},
		},
	}
}

func resourceYandexMessageQueueCreate(d *schema.ResourceData, meta interface{}) error {
	ymqClient, err := newYMQClient(d, meta)
	if err != nil {
		return err
	}

	var name string

	isFifo := d.Get("fifo_queue").(bool)

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		if v, ok := d.GetOk("name_prefix"); ok {
			name = resource.PrefixedUniqueId(v.(string))
		} else {
			name = resource.UniqueId()
		}
		if isFifo {
			name += ".fifo"
		}
	}

	contentBasedDeduplication := d.Get("content_based_deduplication").(bool)

	if !isFifo && contentBasedDeduplication {
		return fmt.Errorf("Content based deduplication can only be set with FIFO queues")
	}

	if isFifo {
		if errors := validateFifoQueueName(name); len(errors) > 0 {
			return fmt.Errorf("Error validating the FIFO queue name: %v", errors)
		}
	} else {
		if errors := validateNonFifoQueueName(name); len(errors) > 0 {
			return fmt.Errorf("Error validating message queue name: %v", errors)
		}
	}

	log.Printf("[INFO] Creating message queue with name %s", name)

	tags := make(map[string]*string)
	for k, v := range d.Get("tags").(map[string]interface{}) {
		tags[k] = aws.String(v.(string))
	}

	req := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Tags:      tags,
	}

	attributes := make(map[string]*string)

	queueResource := *resourceYandexMessageQueue()

	for k, s := range queueResource.Schema {
		if attrKey, ok := sqsQueueAttributeMap[k]; ok {
			if k == "visibility_timeout_seconds" { // NOTE(maxijer@): if visibility_timeout_seconds is set to '0', GetOk will not return its value, so we have to handle it separately
				attributes[attrKey] = aws.String(strconv.Itoa(d.Get(k).(int)))
				continue
			}
			if value, ok := d.GetOk(k); ok {
				switch s.Type {
				case schema.TypeInt:
					attributes[attrKey] = aws.String(strconv.Itoa(value.(int)))
				case schema.TypeBool:
					attributes[attrKey] = aws.String(strconv.FormatBool(value.(bool)))
				default:
					attributes[attrKey] = aws.String(value.(string))
				}
			}

		}
	}

	if len(attributes) > 0 {
		req.Attributes = attributes
	}

	var output *sqs.CreateQueueOutput
	err = resource.Retry(70*time.Second, func() *resource.RetryError {
		var err error
		output, err = ymqClient.CreateQueue(req)
		if err != nil {
			if isAWSSQSErr(err, sqs.ErrCodeQueueDeletedRecently) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error creating message queue: %s", err)
	}

	log.Printf("[INFO] Message queue with name %s was created. Queue url: %s", name, *output.QueueUrl)

	d.SetId(aws.StringValue(output.QueueUrl))

	return resourceYandexMessageQueueReadImpl(d, meta, true)
}

func resourceYandexMessageQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	ymqClient, err := newYMQClient(d, meta)
	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		old, new := d.GetChange("tags")

		oldMap, ok := old.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to parse old tags as map")
		}

		newMap, ok := new.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to parse new tags as map")
		}

		// Collecting tags for deletion
		var tagsToDelete []string
		for tag := range oldMap {
			if _, exists := newMap[tag]; !exists {
				tagsToDelete = append(tagsToDelete, tag)
			}
		}

		// Collecting tags to add/update
		tagsToAdd := make(map[string]string)
		for tag, newValue := range newMap {
			if oldValue, exists := oldMap[tag]; !exists || oldValue.(string) != newValue.(string) {
				tagsToAdd[tag] = newValue.(string)
			}
		}

		// Deleting old tags
		if len(tagsToDelete) > 0 {
			if _, err := ymqClient.UntagQueue(&sqs.UntagQueueInput{
				QueueUrl: aws.String(d.Id()),
				TagKeys:  aws.StringSlice(tagsToDelete),
			}); err != nil {
				return fmt.Errorf("error deleting message queue tags: %s", err)
			}
		}

		// Adding new tags
		if len(tagsToAdd) > 0 {
			if _, err := ymqClient.TagQueue(&sqs.TagQueueInput{
				QueueUrl: aws.String(d.Id()),
				Tags:     aws.StringMap(tagsToAdd),
			}); err != nil {
				return fmt.Errorf("error adding message queue tags: %s", err)
			}
		}
	}

	attributes := make(map[string]*string)

	resource := *resourceYandexMessageQueue()

	for k, s := range resource.Schema {
		if attrKey, ok := sqsQueueAttributeMap[k]; ok {
			if d.HasChange(k) {
				log.Printf("[DEBUG] Updating %s for queue %s", attrKey, d.Id())
				_, n := d.GetChange(k)
				switch s.Type {
				case schema.TypeInt:
					attributes[attrKey] = aws.String(strconv.Itoa(n.(int)))
				case schema.TypeBool:
					attributes[attrKey] = aws.String(strconv.FormatBool(n.(bool)))
				default:
					attributes[attrKey] = aws.String(n.(string))
				}
			}
		}
	}

	if len(attributes) > 0 {
		log.Printf("[INFO] Setting new messsage queue attributes for queue %s", d.Id())

		req := &sqs.SetQueueAttributesInput{
			QueueUrl:   aws.String(d.Id()),
			Attributes: attributes,
		}
		if _, err := ymqClient.SetQueueAttributes(req); err != nil {
			return fmt.Errorf("Error updating message queue attributes: %s", err)
		}

		log.Printf("[INFO] New message queue attributes for queue %s were successfully set", d.Id())
	}

	return resourceYandexMessageQueueReadImpl(d, meta, false)
}

func resourceYandexMessageQueueRead(d *schema.ResourceData, meta interface{}) error {
	return resourceYandexMessageQueueReadImpl(d, meta, false)
}

func resourceYandexMessageQueueDelete(d *schema.ResourceData, meta interface{}) error {
	ymqClient, err := newYMQClient(d, meta)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Delete message queue: %s", d.Id())
	_, err = ymqClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(d.Id()),
	})
	if err == nil {
		log.Printf("[INFO] Message queue %s was successfully deleted", d.Id())
	}
	return err
}

func resourceYandexMessageQueueReadImpl(d *schema.ResourceData, meta interface{}, assumeQueueCreatedRecently bool) error {
	log.Printf("[DEBUG] Reading message queue %s properties", d.Id())
	ymqClient, err := newYMQClient(d, meta)
	if err != nil {
		return err
	}

	var (
		attributeOutput     *sqs.GetQueueAttributesOutput
		queueTagsOutput     *sqs.ListQueueTagsOutput
		listErr, getAttrErr error
	)
	// The first call to resource.Retry for ListQueueTags
	listErr = resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		queueTagsOutput, err = ymqClient.ListQueueTags(&sqs.ListQueueTagsInput{
			QueueUrl: aws.String(d.Id()),
		})

		if err != nil {
			// Queue can be not found immediately after its creation.
			// It can occur in not found or access denied exception.
			if assumeQueueCreatedRecently && (isAWSSQSErr(err, sqs.ErrCodeQueueDoesNotExist) || isAWSSQSErr(err, "AccessDeniedException")) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	// The second call to resource.Retry for GetQueueAttributes
	getAttrErr = resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		attributeOutput, err = ymqClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(d.Id()),
			AttributeNames: []*string{aws.String("All")},
		})

		if err != nil {
			// Queue can be not found immediately after its creation.
			// It can occur in not found or access denied exception.
			if assumeQueueCreatedRecently && (isAWSSQSErr(err, sqs.ErrCodeQueueDoesNotExist) || isAWSSQSErr(err, "AccessDeniedException")) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	// Error handling after both calls
	if listErr != nil || getAttrErr != nil {
		queueNotFoundOrAccessDenied := false

		// Checking for ListQueueTags errors
		if listErr != nil {
			if awsErr, ok := listErr.(awserr.Error); ok {
				if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist || awsErr.Code() == "AccessDeniedException" {
					queueNotFoundOrAccessDenied = true
				}
			}
		}

		// Checking for GetQueueAttributes errors, if not found yet
		if !queueNotFoundOrAccessDenied && getAttrErr != nil {
			if awsErr, ok := getAttrErr.(awserr.Error); ok {
				if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist || awsErr.Code() == "AccessDeniedException" {
					queueNotFoundOrAccessDenied = true
				}
			}
		}

		if queueNotFoundOrAccessDenied {
			d.SetId("")
			log.Printf("[DEBUG] Message queue (%s) not found or access denied", d.Get("name").(string))
			return nil
		}

		// We collect all errors if they are not related to the lack of a queue.
		var errorMessages []string
		if listErr != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("ListQueueTags: %v", listErr))
		}
		if getAttrErr != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("GetQueueAttributes: %v", getAttrErr))
		}
		return fmt.Errorf("errors occurred:\n%v", strings.Join(errorMessages, "\n"))
	}

	name, err := extractNameFromQueueUrl(d.Id())
	if err != nil {
		return err
	}

	// Always set attribute defaults
	d.Set("arn", "")
	d.Set("content_based_deduplication", false)
	d.Set("delay_seconds", 0)
	d.Set("fifo_queue", false)
	d.Set("max_message_size", 262144)
	d.Set("message_retention_seconds", 345600)
	d.Set("name", name)
	d.Set("receive_wait_time_seconds", 0)
	d.Set("redrive_policy", "")
	d.Set("visibility_timeout_seconds", 30)
	d.Set("region_id", defaultYMQRegion)

	if queueTagsOutput != nil {
		if err := d.Set("tags", aws.StringValueMap(queueTagsOutput.Tags)); err != nil {
			return err
		}
	}

	if attributeOutput != nil {
		queueAttributes := aws.StringValueMap(attributeOutput.Attributes)

		if v, ok := queueAttributes[sqs.QueueAttributeNameQueueArn]; ok {
			d.Set("arn", v)
			region, err := regionFromYRN(v)
			if err != nil {
				return err
			}
			d.Set("region_id", region)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameContentBasedDeduplication]; ok && v != "" {
			vBool, err := strconv.ParseBool(v)

			if err != nil {
				return fmt.Errorf("Error parsing content_based_deduplication value (%s) into boolean: %s", v, err)
			}

			d.Set("content_based_deduplication", vBool)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameDelaySeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing delay_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("delay_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameFifoQueue]; ok && v != "" {
			vBool, err := strconv.ParseBool(v)

			if err != nil {
				return fmt.Errorf("Error parsing fifo_queue value (%s) into boolean: %s", v, err)
			}

			d.Set("fifo_queue", vBool)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing kms_data_key_reuse_period_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("kms_data_key_reuse_period_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameKmsMasterKeyId]; ok {
			d.Set("kms_master_key_id", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameMaximumMessageSize]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing max_message_size value (%s) into integer: %s", v, err)
			}

			d.Set("max_message_size", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameMessageRetentionPeriod]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing message_retention_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("message_retention_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNamePolicy]; ok {
			d.Set("policy", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing receive_wait_time_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("receive_wait_time_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameRedrivePolicy]; ok {
			d.Set("redrive_policy", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameVisibilityTimeout]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("Error parsing visibility_timeout_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("visibility_timeout_seconds", vInt)
		}
	}
	return nil
}

func extractNameFromQueueUrl(queue string) (string, error) {
	// Example: https://message-queue.api.cloud.yandex.net/b1g8ad42m6he1ooql78r/dj6000000000qq9v07ol/yet-another-queue
	u, err := url.Parse(queue)
	if err != nil {
		return "", err
	}
	segments := strings.Split(u.Path, "/")
	if len(segments) != 4 {
		return "", fmt.Errorf("Message queue url was not parsed correctly")
	}

	return segments[3], nil
}

func validateQueueName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 80 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 80 characters", k))
	}

	if !regexp.MustCompile(`^[0-9A-Za-z-_]+(\.fifo)?$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("Only alphanumeric characters and hyphens allowed in %q", k))
	}
	return
}

func validateNonFifoQueueName(v interface{}) (errors []error) {
	k := "name"
	value := v.(string)
	if len(value) > 80 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 80 characters", k))
	}

	if !regexp.MustCompile(`^[0-9A-Za-z-_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("Only alphanumeric characters and hyphens allowed in %q", k))
	}
	return
}

func validateFifoQueueName(v interface{}) (errors []error) {
	k := "name"
	value := v.(string)

	if len(value) > 80 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 80 characters", k))
	}

	if !regexp.MustCompile(`^[0-9A-Za-z-_.]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("Only alphanumeric characters and hyphens allowed in %q", k))
	}

	if regexp.MustCompile(`^[^a-zA-Z0-9-_]`).MatchString(value) {
		errors = append(errors, fmt.Errorf("FIFO queue name must start with one of these characters [a-zA-Z0-9-_]: %v", value))
	}

	if !regexp.MustCompile(`\.fifo$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("FIFO queue name should end with \".fifo\": %v", value))
	}

	return
}

func isAWSSQSErr(err error, code string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code
	}
	return false
}

func getKeysForYMQClient(d *schema.ResourceData, meta interface{}) (accessKey, secretKey string, err error) {
	var resourceHasAccessKey, resourcesHasSecretKey bool
	var v interface{}

	if v, resourceHasAccessKey = d.GetOk("access_key"); resourceHasAccessKey {
		accessKey = v.(string)
	}

	if v, resourcesHasSecretKey = d.GetOk("secret_key"); resourcesHasSecretKey {
		secretKey = v.(string)
	}

	if resourceHasAccessKey != resourcesHasSecretKey {
		err = fmt.Errorf("Both access and secret keys should be specified")
		return
	}

	if resourceHasAccessKey { // Keys are in resource
		log.Printf("[DEBUG] Use access and secret keys specified in message queue resource")
	} else { // Keys are in provider
		providerConfig := meta.(*Config)
		if providerConfig.YMQAccessKey == "" || providerConfig.YMQSecretKey == "" {
			err = fmt.Errorf("Message queue access and secret keys should not be empty. They must be specified either in message queue resource or in provider")
			return
		}
		accessKey, secretKey = providerConfig.YMQAccessKey, providerConfig.YMQSecretKey
		log.Printf("[DEBUG] Use access and secret keys specified in provider")
	}

	return
}

func newYMQClientConfigFromKeys(accessKey, secretKey string, providerConfig *Config) *aws.Config {
	return &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(providerConfig.YMQEndpoint),
		Region:      aws.String(providerConfig.Region),
	}
}

func newYMQClientConfig(d *schema.ResourceData, meta interface{}) (config *aws.Config, err error) {
	providerConfig := meta.(*Config)
	accessKey, secretKey, err := getKeysForYMQClient(d, meta)
	if err != nil {
		return
	}
	config = newYMQClientConfigFromKeys(accessKey, secretKey, providerConfig)
	if v, ok := d.GetOk("region_id"); ok {
		log.Printf("[DEBUG] Use custom region: %s", v.(string))
		config.WithRegion(v.(string))
	}
	return
}

func newYMQClientFromConfig(config *aws.Config) (svc *sqs.SQS, err error) {
	newSession, err := session.NewSession(config)
	if err != nil {
		return
	}

	svc = sqs.New(newSession)
	return
}

func newYMQClient(d *schema.ResourceData, meta interface{}) (*sqs.SQS, error) {
	config, err := newYMQClientConfig(d, meta)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] YMQ config: %v", config)

	return newYMQClientFromConfig(config)
}

func regionFromYRN(yrn string) (string, error) {
	// yrn:yc:ymq:ru-central1:21i6v06sqmsaoeon7nus:event-queue
	parts := strings.Split(yrn, ":")
	if len(parts) > 4 {
		return parts[3], nil
	}
	return "", fmt.Errorf("YRN was not parsed correctly: %s", yrn)
}
