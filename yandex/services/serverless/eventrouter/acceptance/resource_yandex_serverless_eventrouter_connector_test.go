package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	eventroutersdk "github.com/yandex-cloud/go-sdk/services/serverless/eventrouter/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const eventrouterConnectorResource = "yandex_serverless_eventrouter_connector.test-connector"

func TestAccEventrouterConnector_yds(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorYds(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "deletion_protection"),
					testYandexEventrouterConnectorContainsLabel(&connector, labelKey, labelValue),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorResource),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "yds.0.database"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "yds.0.stream_name"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "yds.0.consumer"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "yds.0.service_account_id"),
				),
			},
			eventrouterConnectorImportTestStep(),
		},
	})
}

func TestAccEventrouterConnector_ymq(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorYmq(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "deletion_protection"),
					testYandexEventrouterConnectorContainsLabel(&connector, labelKey, labelValue),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorResource),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "ymq.0.queue_arn"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "ymq.0.service_account_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "ymq.0.visibility_timeout"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "ymq.0.batch_size"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "ymq.0.polling_timeout"),
				),
			},
			eventrouterConnectorImportTestStep(),
		},
	})
}

func TestAccEventrouterConnector_timer(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorTimer(name, desc, labelKey, labelValue, "0 45 16 ? * *", "Europe/Moscow", "payload"),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "deletion_protection"),
					testYandexEventrouterConnectorContainsLabel(&connector, labelKey, labelValue),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorResource),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "timer.0.cron_expression"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "timer.0.timezone"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "timer.0.payload"),
				),
			},
			eventrouterConnectorImportTestStep(),
		},
	})
}

func TestAccEventrouterConnector_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var connector eventrouter.Connector
	var connectorUpdated eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	nameUpdated := acctest.RandomWithPrefix("tf-connector-1")
	descUpdated := acctest.RandomWithPrefix("tf-connector-desc-1")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-connector-label-1")
	labelValueUpdated := acctest.RandomWithPrefix("tf-connector-label-value-1")
	queueName := acctest.RandomWithPrefix("tf-queue")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorBasic(name, desc, labelKey, labelValue, queueName),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
				),
			},
			eventrouterConnectorImportTestStep(),
			{
				Config: testYandexEventrouterConnectorBasic(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated, queueName),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connectorUpdated),
					resource.TestCheckResourceAttrWith(eventrouterConnectorResource, "id", func(t *eventrouter.Connector) resource.CheckResourceAttrWithFunc {
						return func(id string) error {
							if id == t.Id {
								return nil
							}
							return errors.New("invalid Event Router connector id")
						}
					}(&connector)),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", nameUpdated),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", descUpdated),
					testYandexEventrouterConnectorContainsLabel(&connectorUpdated, labelKeyUpdated, labelValueUpdated),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorResource),
				),
			},
			eventrouterConnectorImportTestStep(),
		},
	})
}

func (suite testSuite) testYandexEventrouterConnectorDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_serverless_eventrouter_connector" {
			continue
		}

		_, err := testGetEventrouterConnectorByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Event Router connector still exists")
		}
	}

	return nil
}

func eventrouterConnectorImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      eventrouterConnectorResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func (suite testSuite) testYandexEventrouterConnectorExists(name string, connector *eventrouter.Connector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		found, err := testGetEventrouterConnectorByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Event Router connector not found")
		}

		*connector = *found
		return nil
	}
}

func testGetEventrouterConnectorByID(config *providerconfig.Config, ID string) (*eventrouter.Connector, error) {
	req := eventrouter.GetConnectorRequest{
		ConnectorId: ID,
	}

	client := eventroutersdk.NewConnectorClient(config.SDK)

	return client.Get(context.Background(), &req)
}

func testYandexEventrouterConnectorContainsLabel(connector *eventrouter.Connector, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := connector.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexEventrouterConnectorBasic(name, desc, labelKey, labelValue, queueName string) string {
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
}

resource "yandex_message_queue" "test_queue" {
  name = "{{.queue_name}}"
  access_key = yandex_iam_service_account_static_access_key.test_sa_key.access_key
  secret_key = yandex_iam_service_account_static_access_key.test_sa_key.secret_key
  depends_on = [yandex_resourcemanager_folder_iam_member.test-account]
}

resource "yandex_serverless_eventrouter_connector" "test-connector" {
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
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   acceptanceenv.FolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
		"queue_name":  queueName,
	})
	return buf.String()
}

func testYandexEventrouterConnectorYmq(name, desc, labelKey, labelValue string) string {
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
}

resource "yandex_message_queue" "test_queue" {
  name = "{{.name}}-queue"
  access_key = yandex_iam_service_account_static_access_key.test_sa_key.access_key
  secret_key = yandex_iam_service_account_static_access_key.test_sa_key.secret_key
  depends_on = [yandex_resourcemanager_folder_iam_member.test-account]
}

resource "yandex_serverless_eventrouter_connector" "test-connector" {
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
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   acceptanceenv.FolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}

func testYandexEventrouterConnectorYds(name, desc, labelKey, labelValue string) string {
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
  sleep_after = 180
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

resource "yandex_serverless_eventrouter_connector" "test-connector" {
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
    consumer           = tolist(yandex_ydb_topic.test-topic.consumer).0.name
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   acceptanceenv.FolderID(),
		"cloud_id":    acceptanceenv.CloudID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}

func testYandexEventrouterConnectorTimer(name, desc, labelKey, labelValue, cron_expression, timezone, payload string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_serverless_eventrouter_connector" "test-connector" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
  timer {
    cron_expression    = "{{.cron_expression}}"
	timezone		   = "{{.timezone}}"
	payload		       = "{{.payload}}"
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":       acceptanceenv.FolderID(),
		"cloud_id":        acceptanceenv.CloudID(),
		"name":            name,
		"description":     desc,
		"label_key":       labelKey,
		"label_value":     labelValue,
		"cron_expression": cron_expression,
		"timezone":        timezone,
		"payload":         payload,
	})
	return buf.String()
}
