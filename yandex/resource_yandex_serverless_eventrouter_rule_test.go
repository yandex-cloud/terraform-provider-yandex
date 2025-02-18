package yandex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
)

const eventrouterRuleResource = "yandex_serverless_eventrouter_rule.test-rule"

func init() {
	resource.AddTestSweepers("yandex_serverless_eventrouter_rule", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_rule",
		F:    testSweepEventrouterRule,
	})
}

func testSweepEventrouterRule(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListRulesRequest{
		ContainerId: &eventrouter.ListRulesRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}

	it := conf.sdk.Serverless().Eventrouter().Rule().RuleIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepEventrouterRule(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router rule %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterRule(conf *Config, id string) bool {
	return sweepWithRetry(sweepEventrouterRuleOnce, conf, "Event Router rule", id)
}

func sweepEventrouterRuleOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexEventrouterRuleDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Eventrouter().Rule().Delete(ctx, &eventrouter.DeleteRuleRequest{
		RuleId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccEventrouterRule_yds(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleYds(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "yds.0.database"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "yds.0.stream_name"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "yds.0.service_account_id"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_ymq(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleYmq(name, desc, labelKey, labelValue, name),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "ymq.0.queue_arn"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "ymq.0.service_account_id"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_function(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleFunction(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "function.0.function_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "function.0.batch_settings.0.max_count"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "function.0.batch_settings.0.max_bytes"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "function.0.batch_settings.0.cutoff"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_container(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleContainer(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "container.0.container_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "container.0.batch_settings.0.max_count"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "container.0.batch_settings.0.max_bytes"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "container.0.batch_settings.0.cutoff"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_gatewayWsBroadcast(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleGatewayWsBroadcast(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.gateway_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.path"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.service_account_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.batch_settings.0.max_count"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.batch_settings.0.max_bytes"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "gateway_websocket_broadcast.0.batch_settings.0.cutoff"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_logging(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleLogging(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "logging.0.folder_id"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_workflow(t *testing.T) {
	t.Skip("TODO: enable this test when workflow is supported in provider")
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleWorkflow(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "deletion_protection"),
					testYandexEventrouterRuleContainsLabel(&rule, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.workflow_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.service_account_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.gateway_websocket_broadcast.0.batch_settings"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.gateway_websocket_broadcast.0.batch_settings.0.max_count"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.gateway_websocket_broadcast.0.batch_settings.0.max_bytes"),
					resource.TestCheckResourceAttrSet(eventrouterRuleResource, "workflow.0.gateway_websocket_broadcast.0.batch_settings.0.cutoff"),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func TestAccEventrouterRule_update(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	var ruleUpdated eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")
	labelKey := acctest.RandomWithPrefix("tf-rule-label")
	labelValue := acctest.RandomWithPrefix("tf-rule-label-value")

	nameUpdated := acctest.RandomWithPrefix("tf-rule-1")
	descUpdated := acctest.RandomWithPrefix("tf-rule-desc-1")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-rule-label-1")
	labelValueUpdated := acctest.RandomWithPrefix("tf-rule-label-value-1")
	queueName := acctest.RandomWithPrefix("tf-queue")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleYmq(name, desc, labelKey, labelValue, queueName),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &rule),
				),
			},
			eventrouterRuleImportTestStep(),
			{
				Config: testYandexEventrouterRuleYmq(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated, queueName),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleResource, &ruleUpdated),
					resource.TestCheckResourceAttrWith(eventrouterRuleResource, "id", func(t *eventrouter.Rule) resource.CheckResourceAttrWithFunc {
						return func(id string) error {
							if id == t.Id {
								return nil
							}
							return errors.New("invalid Event Router rule id")
						}
					}(&rule)),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "name", nameUpdated),
					resource.TestCheckResourceAttr(eventrouterRuleResource, "description", descUpdated),
					testYandexEventrouterRuleContainsLabel(&ruleUpdated, labelKeyUpdated, labelValueUpdated),
					testAccCheckCreatedAtAttr(eventrouterRuleResource),
				),
			},
			eventrouterRuleImportTestStep(),
		},
	})
}

func testYandexEventrouterRuleDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_serverless_eventrouter_rule" {
			continue
		}

		_, err := testGetEventrouterRuleByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Event Router rule still exists")
		}
	}

	return nil
}

func eventrouterRuleImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      eventrouterRuleResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testYandexEventrouterRuleExists(name string, rule *eventrouter.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetEventrouterRuleByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Event Router rule not found")
		}

		*rule = *found
		return nil
	}
}

func testGetEventrouterRuleByID(config *Config, ID string) (*eventrouter.Rule, error) {
	req := eventrouter.GetRuleRequest{
		RuleId: ID,
	}

	return config.sdk.Serverless().Eventrouter().Rule().Get(context.Background(), &req)
}

func testYandexEventrouterRuleContainsLabel(rule *eventrouter.Rule, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := rule.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexEventrouterRuleYmq(name, desc, labelKey, labelValue, queueName string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "test_sa_key" {
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on = [yandex_resourcemanager_folder_iam_member.test-account]
}

resource "yandex_message_queue" "test_queue" {
  name = "{{.queue_name}}-queue"
  access_key = yandex_iam_service_account_static_access_key.test_sa_key.access_key
  secret_key = yandex_iam_service_account_static_access_key.test_sa_key.secret_key
  depends_on = [yandex_resourcemanager_folder_iam_member.test-account]
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  ymq {
	queue_arn          = yandex_message_queue.test_queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
		"queue_name":  queueName,
	})
	return buf.String()
}

func testYandexEventrouterRuleYds(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_ydb_database_serverless" "test-database" {
  name        = "{{.name}}-ydb-serverless"
  location_id = "ru-central1"
  sleep_after = 60
}

resource "yandex_ydb_topic" "test-topic" {
  database_endpoint = yandex_ydb_database_serverless.test-database.ydb_full_endpoint
  name              = "{{.name}}-topic"

  partitions_count  = 1
  consumer {
    name = "{{.name}}-consumer"
  }
  depends_on = [yandex_ydb_database_serverless.test-database]
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  yds {
    database           = yandex_ydb_database_serverless.test-database.database_path
    stream_name        = yandex_ydb_topic.test-topic.name
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_ydb_topic.test-topic
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"cloud_id":    getExampleCloudID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}

func testYandexEventrouterRuleFunction(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_function" "tf-test" {
  name       = "{{.name}}-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  execution_timeout = "3"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test-account]
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  function {
    function_id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_function.tf-test
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}

func testYandexEventrouterRuleContainer(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_serverless_container" "tf-test" {
  name       = "{{.name}}-container"
  service_account_id = yandex_iam_service_account.test-account.id
  memory = 128
  image {
    url = "{{.container_url}}"
  }
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  container {
    container_id                 = yandex_serverless_container.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_serverless_container.tf-test
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":     getExampleFolderID(),
		"name":          name,
		"description":   desc,
		"label_key":     labelKey,
		"label_value":   labelValue,
		"container_url": serverlessContainerTestImage1,
	})
	return buf.String()
}

func testYandexEventrouterRuleGatewayWsBroadcast(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "api-gateway.websocketBroadcaster"
  sleep_after = 30
}

resource "yandex_api_gateway" "tf-test" {
  name        = "{{.name}}-gateway"
  spec = <<EOF
{{.spec}}EOF
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  gateway_websocket_broadcast {
    gateway_id         = yandex_api_gateway.tf-test.id
    path               = "/test-path"
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_api_gateway.tf-test
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
		"spec":        spec,
	})
	return buf.String()
}

func testYandexEventrouterRuleLogging(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_logging_group" "test-logging-group" {
  name = "{{.name}}-logging-group"
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  logging {
    folder_id = "{{.folder_id}}"
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_logging_group.test-logging-group
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}

func testYandexEventrouterRuleWorkflow(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test-account" {
  folder_id   = "{{.folder_id}}"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_serverless_workflow" "test-workflow" {
  name = "{{.name}}-workflow"
}

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  workflow {
    workflow_id = yandex_serverless_workflow.test-workflow.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [
    yandex_serverless_eventrouter_bus.test-bus,
    yandex_resourcemanager_folder_iam_member.test-account,
    yandex_serverless_workflow.test-workflow
  ]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}
