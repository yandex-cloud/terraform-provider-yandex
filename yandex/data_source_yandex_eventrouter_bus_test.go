package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
)

const eventrouterBusDataSource = "data.yandex_serverless_eventrouter_bus.test-bus"

func TestAccDataSourceEventrouterBus_byID(t *testing.T) {
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusDataSource, &bus),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "bus_id"),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "cloud_id"),
					testAccCheckCreatedAtAttr(eventrouterBusDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceEventrouterBus_byName(t *testing.T) {
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusDataSource, &bus),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "bus_id"),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "cloud_id"),
					testAccCheckCreatedAtAttr(eventrouterBusDataSource),
				),
			},
		},
	})
}

func testYandexEventrouterBusByID(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_serverless_eventrouter_bus" "test-bus" {
  bus_id = yandex_serverless_eventrouter_bus.test-bus.id
}

resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "%s"
  description = "%s"
}
	`, name, desc)
}

func testYandexEventrouterBusByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_serverless_eventrouter_bus" "test-bus" {
  name = yandex_serverless_eventrouter_bus.test-bus.name
}

resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "%s"
  description = "%s"
}
	`, name, desc)
}
