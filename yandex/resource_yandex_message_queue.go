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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateQueueName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"delay_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 900),
			},
			"max_message_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      262144,
				ValidateFunc: validation.IntBetween(1024, 262144),
			},
			"message_retention_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      345600,
				ValidateFunc: validation.IntBetween(60, 1209600),
			},
			"receive_wait_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 20),
			},
			"visibility_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(0, 43200),
			},
			"redrive_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"fifo_queue": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"content_based_deduplication": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultYMQRegion,
				ForceNew: true,
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

			// Computed
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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

	req := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
	}

	attributes := make(map[string]*string)

	queueResource := *resourceYandexMessageQueue()

	for k, s := range queueResource.Schema {
		if attrKey, ok := sqsQueueAttributeMap[k]; ok {
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

	var attributeOutput *sqs.GetQueueAttributesOutput
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
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

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] Found %s", awsErr.Code())
			if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist {
				d.SetId("")
				log.Printf("[DEBUG] Message queue (%s) was not found", d.Get("name").(string))
				return nil
			}
		}
		return err
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
