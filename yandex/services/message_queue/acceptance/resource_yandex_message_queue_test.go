package acceptance

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	awscompatibilitysdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1/awscompatibility"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	iamfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/testfixtures"
	svc_resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"

	"golang.org/x/net/context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	testenv "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_message_queue "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/message_queue/pluginsdk"
)

func TestAccMessageQueue_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes, tags map[string]*string

	resourceName := "yandex_message_queue.queue"
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigWithDefaults(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueDefaultAttributes(&queueAttributes),
				),
			},
			{
				Config: testAccMessageQueueConfigWithOverrides(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueOverrideAttributes(&queueAttributes),
				),
			},
			{
				Config: testAccMessageQueueConfigWithDefaults(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueDefaultAttributes(&queueAttributes),
				),
			},
			{
				Config: testAccMessageQueueConfigWithZeroVisibilityTimeoutSeconds(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueVisibilityZero(&queueAttributes),
				),
			},
			{
				Config: testAccMessageQueueConfigWithDefaults(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueTagsExists(resourceName, &tags),
					testAccCheckMessageQueueDefaultTags(&tags),
				),
			},
		},
	})
}

func TestAccMessageQueue_import(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	cleanup, err := testAccMessageQueueSetTmpKeysForProvider()
	if err != nil {
		t.Fatalf("Error setting tmp credentials: %v", err)
	}
	defer cleanup()

	var queueAttributes map[string]*string

	resourceName := "yandex_message_queue.queue"
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigForImport(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueImportAttributes(&queueAttributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
					"access_key",
					"secret_key",
				},
			},
		},
	})
}

func TestAccMessageQueue_namePrefix(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes map[string]*string

	resourceName := "yandex_message_queue.queue"
	prefix := "acctest-message-queue"
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigWithNamePrefix(prefix, randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueDefaultAttributes(&queueAttributes),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^acctest-message-queue`)),
				),
			},
		},
	})
}

func TestAccMessageQueue_namePrefix_fifo(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes map[string]*string

	resourceName := "yandex_message_queue.queue"
	prefix := "acctest-message-queue"
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueFifoConfigWithNamePrefix(prefix, randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					testAccCheckMessageQueueDefaultAttributes(&queueAttributes),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^acctest-message-queue.*\.fifo$`)),
				),
			},
		},
	})
}

func TestAccMessageQueue_redrivePolicy(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes map[string]*string
	var redriverQueueAttributes map[string]*string

	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigWithRedrive(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists("yandex_message_queue.dead_letter_queue", &queueAttributes),
					testAccCheckMessageQueueDefaultAttributes(&queueAttributes),
					suite.testAccCheckMessageQueueExists("yandex_message_queue.queue", &redriverQueueAttributes),
					testAccCheckMessageQueueRedriverAttributes(&redriverQueueAttributes, &queueAttributes),
				),
			},
		},
	})
}

func TestAccMessageQueue_FIFO(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes map[string]*string

	var randInt int = acctest.RandInt()
	resourceName := "yandex_message_queue.queue"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigWithFIFO(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
				),
			},
		},
	})
}

func TestAccMessageQueue_FIFOExpectNameError(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMessageQueueConfigWithFIFOExpectError(randInt),
				ExpectError: regexp.MustCompile(`Error validating the FIFO queue name`),
			},
		},
	})
}

func TestAccMessageQueue_FIFOWithContentBasedDeduplication(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var queueAttributes map[string]*string

	var randInt int = acctest.RandInt()
	resourceName := "yandex_message_queue.queue"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMessageQueueConfigWithFIFOContentBasedDeduplication(randInt),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckMessageQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "true"),
				),
			},
		},
	})
}

func TestAccMessageQueue_ExpectContentBasedDeduplicationError(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var randInt int = acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckMessageQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccExpectContentBasedDeduplicationError(randInt),
				ExpectError: regexp.MustCompile(`Content based deduplication can only be set with FIFO queues`),
			},
		},
	})
}

func (suite testSuite) testAccNewYMQClientForResource(rs *terraform.ResourceState) (ymqClient *sqs.SQS, err error) {
	cfg, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}
	var accessKey string = rs.Primary.Attributes["access_key"]
	var secretKey string = rs.Primary.Attributes["secret_key"]
	if accessKey == "" || secretKey == "" {

		accessKey, secretKey = cfg.YMQAccessKey, cfg.YMQSecretKey
		if accessKey == "" || secretKey == "" {
			err = fmt.Errorf("Message queue resource has empty access_key or secret_key attribute")
			return
		}
	}
	ymqClient, err = svc_message_queue.NewYMQClientFromConfig(svc_message_queue.NewYMQClientConfigFromKeys(accessKey, secretKey,
		cfg))
	if region, ok := rs.Primary.Attributes["region_id"]; err == nil && ok && region != "" {
		ymqClient.Config.WithRegion(region)
	}
	return
}

func (suite testSuite) testAccCheckMessageQueueDestroy(s *terraform.State) error {
	return suite.testAccCheckMessageQueueDestroyWithProvider(s, suite.Provider())
}

func (suite testSuite) testAccCheckMessageQueueDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	// Create temporary credentials, because credentials
	// that existed during creation of the queue were deleted in terraform destroy procedure.
	providerConfig, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}
	accessKey, secretKey, cleanup, err := createTemporaryStaticAccessKey("editor", providerConfig)
	if err != nil {
		return err
	}
	defer cleanup()
	ymqClient, err := svc_message_queue.NewYMQClientFromConfig(svc_message_queue.NewYMQClientConfigFromKeys(accessKey, secretKey,
		providerConfig))
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_message_queue" {
			continue
		}

		// Check if queue exists by checking for its attributes
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(rs.Primary.ID),
		}
		err := resource.Retry(15*time.Second, func() *resource.RetryError {
			_, err := ymqClient.GetQueueAttributes(params)
			if err != nil {
				if svc_message_queue.IsAWSSQSErr(err, sqs.ErrCodeQueueDoesNotExist) || svc_message_queue.IsAWSSQSErr(err, "AccessDeniedException") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			return resource.RetryableError(fmt.Errorf("Queue %s still exists. Failing!", rs.Primary.ID))
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (suite testSuite) testAccCheckMessageQueueExists(resourceName string, queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		ymqClient, err := suite.testAccNewYMQClientForResource(rs)
		if err != nil {
			return err
		}

		input := &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(rs.Primary.ID),
			AttributeNames: []*string{aws.String("All")},
		}
		output, err := ymqClient.GetQueueAttributes(input)

		if err != nil {
			return err
		}

		*queueAttributes = output.Attributes

		return nil
	}
}

func testAccCheckMessageQueueDefaultAttributes(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if attributes are defaults
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "30" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 30", value)
			}

			if key == "MessageRetentionPeriod" && value != "345600" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 345600", value)
			}

			if key == "MaximumMessageSize" && value != "262144" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 262144", value)
			}

			if key == "DelaySeconds" && value != "0" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 0", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "0" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 0", value)
			}
		}

		return nil
	}
}

func testAccCheckMessageQueueImportAttributes(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "29" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 29", value)
			}

			if key == "MessageRetentionPeriod" && value != "600" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 600", value)
			}

			if key == "MaximumMessageSize" && value != "2049" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 2049", value)
			}

			if key == "DelaySeconds" && value != "303" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 303", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "12" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 12", value)
			}
		}

		return nil
	}
}

func testAccCheckMessageQueueRedriverAttributes(queueAttributes *map[string]*string, sourceQueueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		redrivePolicyPtr, exists := (*queueAttributes)["RedrivePolicy"]
		if !exists {
			return fmt.Errorf("Attribute RedrivePolicy is expected but absent")
		}
		redrivePolicy := aws.StringValue(redrivePolicyPtr)

		var redrivePolicyJson map[string]interface{}
		err := json.Unmarshal([]byte(redrivePolicy), &redrivePolicyJson)
		if err != nil {
			return fmt.Errorf("Failed to parse json from redrive policy: %s", err)
		}

		var maxReceiveCount float64 = redrivePolicyJson["maxReceiveCount"].(float64)
		if maxReceiveCount != 3 {
			return fmt.Errorf("Expected maxReceiveCount == 3, but got %f", maxReceiveCount)
		}

		sourceArnPtr := (*sourceQueueAttributes)["QueueArn"]
		if sourceArnPtr == nil {
			return fmt.Errorf("Attribute QueueArn of source queue is expected but absent")
		}
		sourceArn := aws.StringValue(sourceArnPtr)

		deadLetterTargetArn := redrivePolicyJson["deadLetterTargetArn"].(string)
		if deadLetterTargetArn != sourceArn {
			return fmt.Errorf("Expected deadLetterTargetArn == %s, but got %s", sourceArn, deadLetterTargetArn)
		}

		return nil
	}
}

func testAccCheckMessageQueueVisibilityZero(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if attributes match our overrides
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "0" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 0", value)
			}
			if key == "MessageRetentionPeriod" && value != "86400" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 345600", value)
			}

			if key == "MaximumMessageSize" && value != "2048" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 262144", value)
			}

			if key == "DelaySeconds" && value != "90" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 0", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "10" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 0", value)
			}
		}

		return nil
	}
}

func testAccCheckMessageQueueOverrideAttributes(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if attributes match our overrides
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "60" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 60", value)
			}

			if key == "MessageRetentionPeriod" && value != "86400" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 86400", value)
			}

			if key == "MaximumMessageSize" && value != "2048" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 2048", value)
			}

			if key == "DelaySeconds" && value != "90" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 90", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "10" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 10", value)
			}
		}

		return nil
	}
}

func testAccMessageQueueSetTmpKeysForProvider() (cleanupFunc func(), err error) {
	var endpoint string = os.Getenv("YC_ENDPOINT")
	if endpoint == "" {
		endpoint = common.DefaultEndpoint
	}
	config := providerconfig.Config{
		Endpoint:                       endpoint,
		FolderID:                       acceptanceenv.FolderID(),
		CloudID:                        acceptanceenv.CloudID(),
		Zone:                           os.Getenv("YC_ZONE"),
		Token:                          os.Getenv("YC_TOKEN"),
		ServiceAccountKeyFileOrContent: os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE"),
		Plaintext:                      false,
		Insecure:                       false,
	}

	err = config.InitAndValidate(context.Background(), testenv.TerraformVersion, false)
	if err != nil {
		return
	}

	// Create additional key pair for testing of import.
	accessKey, secretKey, cleanup, err := createTemporaryStaticAccessKey("editor", &config)
	if err != nil {
		err = fmt.Errorf("Failed to create credentials: %s", err)
		return
	}

	var prevAccessKeyValue string = os.Getenv("YC_MESSAGE_QUEUE_ACCESS_KEY")
	var prevSecretKeyValue string = os.Getenv("YC_MESSAGE_QUEUE_SECRET_KEY")

	finalCleanup := func() {
		cleanup()
		os.Setenv("YC_MESSAGE_QUEUE_ACCESS_KEY", prevAccessKeyValue)
		os.Setenv("YC_MESSAGE_QUEUE_SECRET_KEY", prevSecretKeyValue)
	}

	err = os.Setenv("YC_MESSAGE_QUEUE_ACCESS_KEY", accessKey)
	if err != nil {
		finalCleanup()
		err = fmt.Errorf("Failed to set YC_MESSAGE_QUEUE_ACCESS_KEY: %v", err)
		return
	}
	err = os.Setenv("YC_MESSAGE_QUEUE_SECRET_KEY", secretKey)
	if err != nil {
		finalCleanup()
		err = fmt.Errorf("Failed to set YC_MESSAGE_QUEUE_SECRET_KEY: %v", err)
		return
	}

	cleanupFunc = finalCleanup
	return
}

func (suite testSuite) testAccCheckMessageQueueTagsExists(resourceName string, tags *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		ymqClient, err := suite.testAccNewYMQClientForResource(rs)
		if err != nil {
			return err
		}

		input := &sqs.ListQueueTagsInput{
			QueueUrl: aws.String(rs.Primary.ID),
		}
		output, err := ymqClient.ListQueueTags(input)

		if err != nil {
			return err
		}

		*tags = output.Tags

		return nil
	}
}

func testAccCheckMessageQueueDefaultTags(tags *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if tags match our etalon values
		for key, valuePointer := range *tags {
			value := aws.StringValue(valuePointer)
			if key == "env" && value != "prod" {
				return fmt.Errorf("tag 'env' (%s) was not set to 'prod'", value)
			}

			if key == "type" && value != "acctests" {
				return fmt.Errorf("tag 'type' (%s) was not set to 'acctests'", value)
			}
		}

		return nil
	}
}

func testAccMessageQueueConfigWithDefaults(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  region_id   = "ru-central1"
  name        = "message-queue-%d"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  tags = {
    env = "prod"
	type = "acctests"
  }
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigForImport(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  region_id = "ru-central1"
  name      = "message-queue-%d"

  delay_seconds = 303
  max_message_size = 2049
  message_retention_seconds = 600
  receive_wait_time_seconds = 12
  visibility_timeout_seconds = 29
}
`, randInt)
}

func testAccMessageQueueConfigWithNamePrefix(prefix string, randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  region_id   = "ru-central1"
  name_prefix = "%s"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, prefix) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueFifoConfigWithNamePrefix(prefix string, randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name_prefix = "%s"
  fifo_queue  = true

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, prefix) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithOverrides(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name                       = "message-queue-%d"
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithZeroVisibilityTimeoutSeconds(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name                       = "message-queue-%d"
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 0

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithRedrive(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name                       = "tftestqueuq-%d"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key

  redrive_policy = <<EOF
{
  "maxReceiveCount": 3,
  "deadLetterTargetArn": "${yandex_message_queue.dead_letter_queue.arn}"
}
EOF
}

resource "yandex_message_queue" "dead_letter_queue" {
  name = "tfotherqueuq-%d"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithFIFO(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name       = "message-queue-%d.fifo"
  fifo_queue = true

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithFIFOContentBasedDeduplication(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name                        = "message-queue-cbd-%d.fifo"
  fifo_queue                  = true
  content_based_deduplication = true

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccMessageQueueConfigWithFIFOExpectError(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name       = "message-queue-fifo-error-%d"
  fifo_queue = true

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func testAccExpectContentBasedDeduplicationError(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "queue" {
  name                        = "message-queue-cbd-error-%d"
  content_based_deduplication = true

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + iamfixtures.ServiceAccountWithEditor(randInt)
}

func createTemporaryStaticAccessKey(roleID string, config *providerconfig.Config) (accessKey, secretKey string, cleanup func(), err error) {
	serviceAccountClient := iamsdk.NewServiceAccountClient(config.SDK)

	accessKeyClient := awscompatibilitysdk.NewAccessKeyClient(config.SDK)

	op, err := serviceAccountClient.Create(context.Background(), &iam.CreateServiceAccountRequest{
		FolderId: config.FolderID,
		Name:     acctest.RandomWithPrefix("tmp-sa-"),
	})
	if err != nil {
		return
	}

	saID := op.Metadata().ServiceAccountId

	_, err = op.Wait(context.Background())
	if err != nil {
		return
	}

	deleteSa := func() {
		// Use a folder updater in a similar way iam resources do to prevent modifying folder access binding while
		// service account is being deleted.
		updater, err := svc_resource_manager.NewFolderIamUpdaterFromFolderID(config.FolderID, config)
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
			return
		}

		mutexKey := updater.GetMutexKey()
		sdkcommon.MutexKV.Lock(mutexKey)
		defer sdkcommon.MutexKV.Unlock(mutexKey)

		op, err := serviceAccountClient.Delete(context.Background(), &iam.DeleteServiceAccountRequest{
			ServiceAccountId: saID,
		})
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
			return
		}

		_, err = op.Wait(context.Background())
		if err != nil {
			log.Printf("[WARN] error deleting temporary service account: %s", err)
		}
	}

	createKey := func() (*awscompatibility.CreateAccessKeyResponse, error) {
		client := resourcemanagersdk.NewFolderClient(config.SDK)

		updateOp, updateErr := client.UpdateAccessBindings(context.Background(), &access.UpdateAccessBindingsRequest{
			ResourceId: config.FolderID,
			AccessBindingDeltas: []*access.AccessBindingDelta{
				{
					Action: access.AccessBindingAction_ADD,
					AccessBinding: &access.AccessBinding{
						RoleId: roleID,
						Subject: &access.Subject{
							Id:   saID,
							Type: "serviceAccount",
						},
					},
				},
			},
		})
		if updateErr != nil {
			return nil, updateErr
		}

		_, updateErr = updateOp.Wait(context.Background())
		if updateErr != nil {
			return nil, updateErr
		}

		sak, err := accessKeyClient.Create(context.Background(), &awscompatibility.CreateAccessKeyRequest{
			ServiceAccountId: saID,
		})
		if err != nil {
			return nil, err
		}

		return sak, err
	}

	sak, err := createKey()
	if err != nil {
		deleteSa()
		return
	}

	accessKey = sak.AccessKey.KeyId
	secretKey = sak.Secret
	cleanup = func() {
		_, err := accessKeyClient.Delete(context.Background(), &awscompatibility.DeleteAccessKeyRequest{
			AccessKeyId: sak.AccessKey.Id,
		})
		if err != nil {
			log.Printf("[WARN] error deleting temporary access key: %s", err)
		}

		deleteSa()
	}
	return
}
