package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	eventroutersdk "github.com/yandex-cloud/go-sdk/services/serverless/eventrouter/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const eventrouterBusDataSource = "data.yandex_serverless_eventrouter_bus.test-bus"

func TestAccDataSourceEventrouterBus_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterBusExists(eventrouterBusDataSource, &bus),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "bus_id"),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "cloud_id"),
					statecheck.CheckCreatedAtAttr(eventrouterBusDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceEventrouterBus_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexEventrouterBusExists(eventrouterBusDataSource, &bus),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "bus_id"),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusDataSource, "cloud_id"),
					statecheck.CheckCreatedAtAttr(eventrouterBusDataSource),
				),
			},
		},
	})
}

func (suite testSuite) testYandexEventrouterBusDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

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

func (suite testSuite) testYandexEventrouterBusExists(name string, bus *eventrouter.Bus) resource.TestCheckFunc {
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

func testGetEventrouterBusByID(config *providerconfig.Config, ID string) (*eventrouter.Bus, error) {
	req := eventrouter.GetBusRequest{
		BusId: ID,
	}

	client := eventroutersdk.NewBusClient(config.SDK)

	return client.Get(context.Background(), &req)
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
