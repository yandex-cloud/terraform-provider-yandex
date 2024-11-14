package yandex

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
)

const eventrouterRuleDataSource = "data.yandex_serverless_eventrouter_rule.test-rule"

func TestAccDataSourceEventrouterRule_byID(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleDataSource, &rule),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "rule_id"),
					resource.TestCheckResourceAttr(eventrouterRuleDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "cloud_id"),
					testAccCheckCreatedAtAttr(eventrouterRuleDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceEventrouterRule_byName(t *testing.T) {
	t.Parallel()

	var rule eventrouter.Rule
	name := acctest.RandomWithPrefix("tf-rule")
	desc := acctest.RandomWithPrefix("tf-rule-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexEventrouterRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterRuleByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterRuleExists(eventrouterRuleDataSource, &rule),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "rule_id"),
					resource.TestCheckResourceAttr(eventrouterRuleDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterRuleDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterRuleDataSource, "cloud_id"),
					testAccCheckCreatedAtAttr(eventrouterRuleDataSource),
				),
			},
		},
	})
}

func testYandexEventrouterRuleByID(name string, desc string) string {
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

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  ymq {
	queue_arn          = yandex_message_queue.test_queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}

data "yandex_serverless_eventrouter_rule" "test-rule" {
  rule_id = yandex_serverless_eventrouter_rule.test-rule.id
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
	})
	return buf.String()
}

func testYandexEventrouterRuleByName(name string, desc string) string {
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

resource "yandex_serverless_eventrouter_rule" "test-rule" {
  bus_id      = yandex_serverless_eventrouter_bus.test-bus.id
  name        = "{{.name}}"
  description = "{{.description}}"
  ymq {
	queue_arn          = yandex_message_queue.test_queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
  depends_on = [yandex_serverless_eventrouter_bus.test-bus]
}

data "yandex_serverless_eventrouter_rule" "test-rule" {
  name = yandex_serverless_eventrouter_rule.test-rule.name
}
`))
	buf := &bytes.Buffer{}
	_ = tmpl.Execute(buf, map[string]interface{}{
		"folder_id":   getExampleFolderID(),
		"name":        name,
		"description": desc,
	})
	return buf.String()
}
