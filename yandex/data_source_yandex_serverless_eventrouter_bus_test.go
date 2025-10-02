package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
)

const eventrouterBusDataSource = "data.yandex_serverless_eventrouter_bus.test-bus"

func TestAccDataSourceEventrouterBus_byID(t *testing.T) {
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testYandexEventrouterBusDestroy,
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testYandexEventrouterBusDestroy,
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

func testYandexEventrouterBusDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_serverless_eventrouter_bus" {
			continue
		}

		_, err := testGetEventrouterBusByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Event Router bus still exists")
		}
	}

	return nil
}

func testYandexEventrouterBusExists(name string, bus *eventrouter.Bus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetEventrouterBusByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Event Router bus not found")
		}

		*bus = *found
		return nil
	}
}

func testGetEventrouterBusByID(config *Config, ID string) (*eventrouter.Bus, error) {
	req := eventrouter.GetBusRequest{
		BusId: ID,
	}

	return config.sdk.Serverless().Eventrouter().Bus().Get(context.Background(), &req)
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
