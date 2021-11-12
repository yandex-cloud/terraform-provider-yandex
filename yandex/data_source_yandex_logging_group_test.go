package yandex

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

const yandexLoggingGroupDataSource = "data.yandex_logging_group.test-logging-group"

func TestAccDataSourceYandexLoggingGroup_byID(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "group_id"),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", name),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "retention_period"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
					testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLoggingGroup_byName(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "group_id"),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", name),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "retention_period"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
					testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLoggingGroup_full(t *testing.T) {
	var group logging.LogGroup
	params := testYandexLoggingGroupParameters{
		name:            acctest.RandomWithPrefix("tf-yandex-logging-group"),
		desc:            acctest.RandomWithPrefix("tf-yandex-logging-group-desc"),
		labelKey:        acctest.RandomWithPrefix("tf-yandex-logging-group-label"),
		labelValue:      acctest.RandomWithPrefix("tf-yandex-logging-group-label-value"),
		retentionPeriod: time.Hour + time.Duration(rand.Uint32())*time.Nanosecond,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", params.name),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", params.desc),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "retention_period", params.retentionPeriod.String()),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
					testYandexLoggingGroupContainsLabel(&group, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
				),
			},
		},
	})
}

func testYandexLoggingGroupByID(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  group_id = "${yandex_logging_group.test-logging-group.id}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
}`, name, desc)
}

func testYandexLoggingGroupByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  name = "${yandex_logging_group.test-logging-group.name}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testYandexLoggingGroupDataSource(params testYandexLoggingGroupParameters) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  group_id = "${yandex_logging_group.test-logging-group.id}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  retention_period = "%s"
}
`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.retentionPeriod,
	)
}
