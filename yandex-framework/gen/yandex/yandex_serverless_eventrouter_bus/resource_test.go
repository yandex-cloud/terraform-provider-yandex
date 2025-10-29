package yandex_serverless_eventrouter_bus_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	eventrouterv1sdk "github.com/yandex-cloud/go-sdk/services/serverless/eventrouter/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const eventrouterBusResource = "yandex_serverless_eventrouter_bus.test-bus"

func init() {
	resource.AddTestSweepers("yandex_serverless_eventrouter_bus", &resource.Sweeper{
		Name: "yandex_serverless_eventrouter_bus",
		F:    testSweepEventrouterBus,
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepEventrouterBus(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &eventrouter.ListBusesRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	resp, err := eventrouterv1sdk.NewBusClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting buses: %s", err)
	}
	result := &multierror.Error{}
	for _, b := range resp.Buses {
		id := b.GetId()
		if !test.SweepWithRetry(sweepEventrouterBusOnce, conf, "Event Router bus", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep sweep Event Router bus %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepEventrouterBusOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := eventrouterv1sdk.NewBusClient(conf.SDKv2).Delete(ctx, &eventrouter.DeleteBusRequest{
		BusId: id,
	})
	_, err = op.Wait(ctx)
	return err
}

func TestAccEventrouterBus_UpgradeFromSDKv2(t *testing.T) {
	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")
	labelKey := acctest.RandomWithPrefix("tf-bus-label")
	labelValue := acctest.RandomWithPrefix("tf-bus-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testYandexEventrouterBusDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testYandexEventrouterBusBasic(name, desc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexEventrouterBusExists(eventrouterBusResource, &bus),
					resource.TestCheckResourceAttr(eventrouterBusResource, "name", name),
					resource.TestCheckResourceAttr(eventrouterBusResource, "description", desc),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "folder_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(eventrouterBusResource, "deletion_protection"),
					testYandexEventrouterBusContainsLabel(&bus, labelKey, labelValue),
					test.AccCheckCreatedAtAttr(eventrouterBusResource),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testYandexEventrouterBusBasic(name, desc, labelKey, labelValue),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccEventrouterBus_basic(t *testing.T) {
	var bus eventrouter.Bus
	name := acctest.RandomWithPrefix("tf-bus")
	desc := acctest.RandomWithPrefix("tf-bus-desc")
	labelKey := acctest.RandomWithPrefix("tf-bus-label")
	labelValue := acctest.RandomWithPrefix("tf-bus-label-value")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testYandexEventrouterBusDestroy,
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
					test.AccCheckCreatedAtAttr(eventrouterBusResource),
				),
			},
			eventrouterBusImportTestStep(),
		},
	})
}

func TestAccEventrouterBus_update(t *testing.T) {
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testYandexEventrouterBusDestroy,
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
					test.AccCheckCreatedAtAttr(eventrouterBusResource),
				),
			},
			eventrouterBusImportTestStep(),
		},
	})
}

func testYandexEventrouterBusDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

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

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

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

func testGetEventrouterBusByID(conf provider_config.Config, ID string) (*eventrouter.Bus, error) {
	req := eventrouter.GetBusRequest{
		BusId: ID,
	}

	return eventrouterv1sdk.NewBusClient(conf.SDKv2).Get(context.Background(), &req)
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
		"folder_id":   test.GetExampleFolderID(),
		"name":        name,
		"description": desc,
		"label_key":   labelKey,
		"label_value": labelValue,
	})
	return buf.String()
}
