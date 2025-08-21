package yandex_cloudregistry_registry_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const yandexCloudRegistryDefaultTimeout = 15 * time.Minute

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_cloudregistry_registry", &resource.Sweeper{
		Name:         "yandex_cloudregistry_registry",
		F:            testSweepCloudRegistry,
		Dependencies: []string{},
	})
}

func testSweepCloudRegistry(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cloudregistry.ListRegistriesRequest{FolderId: test.GetExampleFolderID()}
	it := conf.SDK.CloudRegistry().Registry().RegistryIterator(context.Background(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepCloudRegistry(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Cloud Registry %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCloudRegistry(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepcloudRegistryOnce, conf, "Cloud Registry", id)
}

func sweepcloudRegistryOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexCloudRegistryDefaultTimeout)
	defer cancel()

	op, err := conf.SDK.CloudRegistry().Registry().Delete(ctx, &cloudregistry.DeleteRegistryRequest{
		RegistryId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

//revive:disable:var-naming
func TestAccCloudRegistry_basic(t *testing.T) {

	registryName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var registry cloudregistry.Registry
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, cloudregistry.Registry_Kind_name[3], cloudregistry.Registry_Type_name[1], "my-value-for-tag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					test.AccCheckCreatedAtAttr("yandex_cloudregistry_registry.foobar"),
					testAccCheckCloudRegistryName(&registry, registryName),
					testAccCheckCloudRegistryContainsLabel(&registry, "test-label", "my-value-for-tag"),
					testAccCheckCloudRegistryStatus(&registry, "active"),
				),
			},
		},
	})
}

func TestAccCloudRegistry_updateNameAndLabels(t *testing.T) {

	var registry cloudregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := test.GetExampleFolderID()
	var registryID string
	var afterUpdateRegistryID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					func(s *terraform.State) error {
						registryID = registry.Id
						return nil
					},
				),
			},
			{
				Config: testAccCloudRegistry_update("new-registry-name", folderID, "DOCKER", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_cloudregistry_registry.foobar", "id", &registry.Id),
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar", "name", "new-registry-name"),
					testAccCheckCloudRegistryName(&registry, "new-registry-name"),
					testAccCheckCloudRegistryContainsLabel(&registry, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckCloudRegistryContainsLabel(&registry, "new-field", "only-shows-up-when-updated"),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar",
						"labels.empty-label", "oh-look-theres-a-label-now"),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar",
						"labels.new-field", "only-shows-up-when-updated"),
					testAccCheckCloudRegistryDoesNotContainLabel(&registry, "test-label"),
					func(s *terraform.State) error {
						afterUpdateRegistryID = registry.Id
						return nil
					},
					testAccCheckCloudRegistyIdsEqual(&registryID, &afterUpdateRegistryID),
				),
			},
			{
				ResourceName:      "yandex_cloudregistry_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistry_updateOnlyName(t *testing.T) {

	var registry cloudregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := test.GetExampleFolderID()
	var registryID string
	var afterUpdateRegistryID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					testAccCheckCloudRegistryLabel(&registry, "test-label", "my-init-value"),
					func(s *terraform.State) error {
						registryID = registry.Id
						return nil
					},
				),
			},
			{
				Config: testAccCloudRegistry_basic("new-registry-name", folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_cloudregistry_registry.foobar", "id", &registry.Id),
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar", "name", "new-registry-name"),
					testAccCheckCloudRegistryContainsLabel(&registry, "test-label", "my-init-value"),
					testAccCheckCloudRegistryName(&registry, "new-registry-name"),
					func(s *terraform.State) error {
						afterUpdateRegistryID = registry.Id
						return nil
					},
					testAccCheckCloudRegistyIdsEqual(&registryID, &afterUpdateRegistryID),
				),
			},
			{
				ResourceName:      "yandex_cloudregistry_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistry_updateOnlyLabels(t *testing.T) {

	var registry cloudregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", strings.ToLower(acctest.RandString(10)))
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
				),
			},
			{
				Config: testAccCloudRegistry_update(registryName, folderID, "DOCKER", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_cloudregistry_registry.foobar", "id", &registry.Id),
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					testAccCheckCloudRegistryContainsLabel(&registry, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckCloudRegistryContainsLabel(&registry, "new-field", "only-shows-up-when-updated"),
					testAccCheckCloudRegistryDoesNotContainLabel(&registry, "test-label"),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar",
						"labels.empty-label", "oh-look-theres-a-label-now"),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar",
						"labels.new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_cloudregistry_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistry_updateOnlyDescription(t *testing.T) {

	var registry cloudregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := test.GetExampleFolderID()
	var registryID string
	var afterUpdateRegistryID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					testAccCheckCloudRegistryLabel(&registry, "test-label", "my-init-value"),
				),
			},
			{
				Config: testAccCloudRegistry_updateDescription(registryName, folderID, "DOCKER", "LOCAL", "new-description"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_cloudregistry_registry.foobar", "id", &registry.Id),
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar", "description", "new-description"),
					testAccCheckCloudRegistryDescription(&registry, "new-description"),
					testAccCheckCloudRegistyIdsEqual(&registryID, &afterUpdateRegistryID),
				),
			},
			{
				ResourceName:      "yandex_cloudregistry_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistry_updateLabelValue(t *testing.T) {

	var registry cloudregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", strings.ToLower(acctest.RandString(10)))
	folderID := test.GetExampleFolderID()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
				),
			},
			{
				Config: testAccCloudRegistry_basic(registryName, folderID, "DOCKER", "LOCAL", "my-new-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_cloudregistry_registry.foobar", "id", &registry.Id),
					testAccCheckCloudRegistryExists("yandex_cloudregistry_registry.foobar", &registry),
					testAccCheckCloudRegistryContainsLabel(&registry, "test-label", "my-new-value"),
					resource.TestCheckResourceAttr("yandex_cloudregistry_registry.foobar",
						"labels.test-label", "my-new-value"),
				),
			},
			{
				ResourceName:      "yandex_cloudregistry_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudRegistryDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cloudregistry_registry" {
			continue
		}

		_, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Cloud Registry still exists")
		}
	}

	return nil
}

func testAccCheckCloudRegistryExists(n string, registry *cloudregistry.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Cloud Registry %s not found", n)
		}

		*registry = *found
		return nil
	}
}

func testAccCheckCloudRegistryName(registry *cloudregistry.Registry, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registry.Name != name {
			return fmt.Errorf("Wrong Cloud Registry name: expected '%s' got '%s'", name, registry.Name)
		}
		return nil
	}
}

func testAccCheckCloudRegistryLabel(registry *cloudregistry.Registry, key_label string, label string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registry.Labels[key_label] != label {
			return fmt.Errorf("Wrong Cloud Registry test-label: expected '%s' got '%s'", label, registry.Labels[key_label])
		}
		return nil
	}
}

func testAccCheckCloudRegistryStatus(registry *cloudregistry.Registry, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		registryStatus := strings.ToLower(registry.Status.String())
		if registryStatus != status {
			return fmt.Errorf("Wrong Cloud Registry status: expected '%s' got '%s'", status, registryStatus)
		}
		return nil
	}
}

func testAccCheckCloudRegistryContainsLabel(registry *cloudregistry.Registry, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := registry.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckCloudRegistryDescription(registry *cloudregistry.Registry, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registry.Description != description {
			return fmt.Errorf("Wrong Cloud Registry description: expected '%s' got '%s'", description, registry.Description)
		}
		return nil
	}
}

func testAccCheckCloudRegistryDoesNotContainLabel(registry *cloudregistry.Registry, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v, ok := registry.Labels[key]; ok {
			return fmt.Errorf("Expected no label for key '%s' but found one with value '%s'", key, v)
		}

		return nil
	}
}

func testAccCheckCloudRegistyIdsEqual(registryID *string, afterUpdateRegistryID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *registryID != *afterUpdateRegistryID {
			return fmt.Errorf("Cloud Registry id has changed: before '%s', after update '%s'", *registryID, *afterUpdateRegistryID)
		}

		return nil
	}
}

func testAccCloudRegistry_update(name, folderID, kind, typeName string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"
  kind      = "%s"
  type		= "%s"
  
  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, name, folderID, kind, typeName)
}

func testAccCloudRegistry_basic(name, folderID, kind, typeName, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"
  kind      = "%s"
  type		= "%s"

  labels = {
    test-label = "%s"
  }
}
`, name, folderID, kind, typeName, labelValue)
}

func testAccCloudRegistry_updateDescription(name, folderID, kind, typeName, description string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"
  kind      = "%s"
  type		= "%s"
  description = "%s"
  
}
`, name, folderID, kind, typeName, description)
}
