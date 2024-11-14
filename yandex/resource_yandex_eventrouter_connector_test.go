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

const eventrouterConnectorResource = "yandex_serverless_eventrouter_connector.test-connector"

func init() {
	resource.AddTestSweepers("yandex_serverless_eventrouter_connector", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_connector",
		F:    testSweepEventrouterConnector,
	})
}

func testSweepEventrouterConnector(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListConnectorsRequest{
		ContainerId: &eventrouter.ListConnectorsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}

	it := conf.sdk.Serverless().Eventrouter().Connector().ConnectorIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepEventrouterConnector(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router connector %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterConnector(conf *Config, id string) bool {
	return sweepWithRetry(sweepEventrouterConnectorOnce, conf, "Event Router connector", id)
}

func sweepEventrouterConnectorOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexEventrouterConnectorDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Eventrouter().Connector().Delete(ctx, &eventrouter.DeleteConnectorRequest{
		ConnectorId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccEventrouterConnector_yds(t *testing.T) {
	t.Parallel()

	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorYds(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "deletion_protection"),
					testYandexEventrouterConnectorContainsLabel(&connector, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterConnectorResource),
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
	t.Parallel()

	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")
	labelKey := acctest.RandomWithPrefix("tf-connector-label")
	labelValue := acctest.RandomWithPrefix("tf-connector-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorYmq(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorResource, "deletion_protection"),
					testYandexEventrouterConnectorContainsLabel(&connector, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterConnectorResource),
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

func TestAccEventrouterConnector_update(t *testing.T) {
	t.Parallel()

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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorBasic(name, desc, labelKey, labelValue, queueName),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connector),
				),
			},
			eventrouterConnectorImportTestStep(),
			{
				Config: testYandexEventrouterConnectorBasic(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated, queueName),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterConnectorExists(eventrouterConnectorResource, &connectorUpdated),
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
					testAccCheckCreatedAtAttr(eventrouterConnectorResource),
				),
			},
			eventrouterConnectorImportTestStep(),
		},
	})
}

func testYandexEventrouterConnectorDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

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

func testYandexEventrouterConnectorExists(name string, connector *eventrouter.Connector) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

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

func testGetEventrouterConnectorByID(config *Config, ID string) (*eventrouter.Connector, error) {
	req := eventrouter.GetConnectorRequest{
		ConnectorId: ID,
	}

	return config.sdk.Serverless().Eventrouter().Connector().Get(context.Background(), &req)
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
		"folder_id":   getExampleFolderID(),
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
		"folder_id":   getExampleFolderID(),
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
  sleep_after = 60
}

resource "yandex_ydb_topic" "test-topic" {
  database_endpoint = yandex_ydb_database_serverless.test-database.ydb_full_endpoint
  name              = "{{.name}}-topic"

  partitions_count  = 1
  consumer {
    name = "{{.name}}-consumer"
  }
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
		"folder_id":   getExampleFolderID(),
		"cloud_id":    getExampleCloudID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}
