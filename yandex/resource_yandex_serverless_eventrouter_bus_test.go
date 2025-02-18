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

const eventrouterBusResource = "yandex_serverless_eventrouter_bus.test-bus"

func init() {
	resource.AddTestSweepers("yandex_serverless_eventrouter_bus", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_bus",
		F:    testSweepEventrouterBus,
	})
}

func testSweepEventrouterBus(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListBusesRequest{FolderId: conf.FolderID}
	it := conf.sdk.Serverless().Eventrouter().Bus().BusIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepEventrouterBus(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router bus %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterBus(conf *Config, id string) bool {
	return sweepWithRetry(sweepEventrouterBusOnce, conf, "Event Router bus", id)
}

func sweepEventrouterBusOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexEventrouterBusDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Eventrouter().Bus().Delete(ctx, &eventrouter.DeleteBusRequest{
		BusId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccEventrouterBus_basic(t *testing.T) {
	t.Parallel()

	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")
	labelKey := acctest.RandomWithPrefix("tf-bus-label")
	labelValue := acctest.RandomWithPrefix("tf-bus-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusBasic(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusResource, &bus),
					resource.TestCheckResourceAttr(eventrouterBusResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "deletion_protection"),
					testYandexEventrouterBusContainsLabel(&bus, labelKey, labelValue),
					testAccCheckCreatedAtAttr(eventrouterBusResource),
				),
			},
			eventrouterBusImportTestStep(),
		},
	})
}

func TestAccEventrouterBus_update(t *testing.T) {
	t.Parallel()

	var bus eventrouter.Bus
	var busUpdated eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")
	labelKey := acctest.RandomWithPrefix("tf-bus-label")
	labelValue := acctest.RandomWithPrefix("tf-bus-label-value")

	nameUpdated := acctest.RandomWithPrefix("tf-bus-1")
	descUpdated := acctest.RandomWithPrefix("tf-bus-desc-1")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-bus-label-1")
	labelValueUpdated := acctest.RandomWithPrefix("tf-bus-label-value-1")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexEventrouterBusBasic(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusResource, &bus),
				),
			},
			eventrouterBusImportTestStep(),
			{
				Config: testYandexEventrouterBusBasic(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusResource, &busUpdated),
					resource.TestCheckResourceAttrWith(eventrouterBusResource, "id", func(t *eventrouter.Bus) resource.CheckResourceAttrWithFunc {
						return func(id string) error {
							if id == t.Id {
								return nil
							}
							return errors.New("invalid Event Router bus id")
						}
					}(&bus)),
					resource.TestCheckResourceAttr(eventrouterBusResource, "name", nameUpdated),
					resource.TestCheckResourceAttr(eventrouterBusResource, "description", descUpdated),
					testYandexEventrouterBusContainsLabel(&busUpdated, labelKeyUpdated, labelValueUpdated),
					testAccCheckCreatedAtAttr(eventrouterBusResource),
				),
			},
			eventrouterBusImportTestStep(),
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

func eventrouterBusImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      eventrouterBusResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
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

func testYandexEventrouterBusContainsLabel(bus *eventrouter.Bus, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := bus.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexEventrouterBusBasic(name, desc, labelKey, labelValue string) string {
	tmpl := template.Must(template.New("tf").Parse(`
resource "yandex_serverless_eventrouter_bus" "test-bus" {
  name        = "{{.name}}"
  description = "{{.description}}"
  folder_id   = "{{.folder_id}}"
  labels = {
    {{.label_key}}          = "{{.label_value}}"
    empty-label = ""
  }
}`))
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
