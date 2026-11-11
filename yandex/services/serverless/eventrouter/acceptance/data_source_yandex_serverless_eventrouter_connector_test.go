package acceptance

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const eventrouterConnectorDataSource = "data.yandex_serverless_eventrouter_connector.test-connector"

func TestAccDataSourceEventrouterConnector_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorDataSource, &connector),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "connector_id"),
					resource.TestCheckResourceAttr(eventrouterConnectorDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "cloud_id"),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceEventrouterConnector_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var connector eventrouter.Connector
	name := acctest.RandomWithPrefix("tf-connector")
	desc := acctest.RandomWithPrefix("tf-connector-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterConnectorByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterConnectorExists(eventrouterConnectorDataSource, &connector),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "connector_id"),
					resource.TestCheckResourceAttr(eventrouterConnectorDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterConnectorDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterConnectorDataSource, "cloud_id"),
					statecheck.CheckCreatedAtAttr(eventrouterConnectorDataSource),
				),
			},
		},
	})
}

func testYandexEventrouterConnectorByID(name string, desc string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
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
  depends_on = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_serverless_eventrouter_connector" "test-connector" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  ymq {
	queue_arn          = yandex_message_queue.test_queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}

data "yandex_serverless_eventrouter_connector" "test-connector" {
  connector_id = yandex_serverless_eventrouter_connector.test-connector.id
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   acceptanceenv.FolderID(),
		"name":        name,
		"description": desc,
	})
	return buf.String()
}

func testYandexEventrouterConnectorByName(name string, desc string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
}

resource "yandex_iam_service_account" "test-account" {
  name = "{{.name}}-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
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
  depends_on = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_serverless_eventrouter_connector" "test-connector" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  ymq {
	queue_arn          = yandex_message_queue.test_queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}

data "yandex_serverless_eventrouter_connector" "test-connector" {
  name = yandex_serverless_eventrouter_connector.test-connector.name
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   acceptanceenv.FolderID(),
		"name":        name,
		"description": desc,
	})
	return buf.String()
}
