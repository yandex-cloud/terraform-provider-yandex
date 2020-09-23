package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

//revive:disable:var-naming
func TestAccContainerRegistry_basic(t *testing.T) {
	t.Parallel()

	registryName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	var registry containerregistry.Registry
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-value-for-tag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
					testAccCheckCreatedAtAttr("yandex_container_registry.foobar"),
					testAccCheckContainerRegistryName(&registry, registryName),
					testAccCheckContainerRegistryContainsLabel(&registry, "test_label", "my-value-for-tag"),
					testAccCheckContainerRegistryStatus(&registry, "active"),
				),
			},
		},
	})
}

func TestAccContainerRegistry_updateNameAndLabels(t *testing.T) {
	t.Parallel()

	var registry containerregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := getExampleFolderID()
	var registryID string
	var afterUpdateRegistryID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
				),
			},
			{
				Config: testAccContainerRegistry_update("new-registry-name", folderID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_container_registry.foobar", "id", &registry.Id),
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar", "name", "new-registry-name"),
					testAccCheckContainerRegistryName(&registry, "new-registry-name"),
					testAccCheckContainerRegistryContainsLabel(&registry, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckContainerRegistryContainsLabel(&registry, "new-field", "only-shows-up-when-updated"),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar",
						"labels.empty-label", "oh-look-theres-a-label-now"),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar",
						"labels.new-field", "only-shows-up-when-updated"),
					testAccCheckContainerRegistryDoesNotContainLabel(&registry, "test_label"),
					testAccCheckRegistyIdsEqual(&registryID, &afterUpdateRegistryID),
				),
			},
			{
				ResourceName:      "yandex_container_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccContainerRegistry_updateOnlyName(t *testing.T) {
	t.Parallel()

	var registry containerregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := getExampleFolderID()
	var registryID string
	var afterUpdateRegistryID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
				),
			},
			{
				Config: testAccContainerRegistry_basic("new-registry-name", folderID, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_container_registry.foobar", "id", &registry.Id),
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar", "name", "new-registry-name"),
					testAccCheckContainerRegistryName(&registry, "new-registry-name"),
					testAccCheckRegistyIdsEqual(&registryID, &afterUpdateRegistryID),
				),
			},
			{
				ResourceName:      "yandex_container_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccContainerRegistry_updateOnlyLabels(t *testing.T) {
	t.Parallel()

	var registry containerregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
				),
			},
			{
				Config: testAccContainerRegistry_update(registryName, folderID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_container_registry.foobar", "id", &registry.Id),
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
					testAccCheckContainerRegistryContainsLabel(&registry, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckContainerRegistryContainsLabel(&registry, "new-field", "only-shows-up-when-updated"),
					testAccCheckContainerRegistryDoesNotContainLabel(&registry, "test_label"),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar",
						"labels.empty-label", "oh-look-theres-a-label-now"),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar",
						"labels.new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_container_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccContainerRegistry_updateLabelValue(t *testing.T) {
	t.Parallel()

	var registry containerregistry.Registry
	registryName := fmt.Sprintf("tf-test-update-%s", acctest.RandString(10))
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckContainerRegisterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-init-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
				),
			},
			{
				Config: testAccContainerRegistry_basic(registryName, folderID, "my-new-value"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr("yandex_container_registry.foobar", "id", &registry.Id),
					testAccCheckContainerRegistryExists("yandex_container_registry.foobar", &registry),
					testAccCheckContainerRegistryContainsLabel(&registry, "test_label", "my-new-value"),
					resource.TestCheckResourceAttr("yandex_container_registry.foobar",
						"labels.test_label", "my-new-value"),
				),
			},
			{
				ResourceName:      "yandex_container_registry.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckContainerRegisterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_registry" {
			continue
		}

		_, err := config.sdk.ContainerRegistry().Registry().Get(context.Background(), &containerregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex.Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Registry still exists")
		}
	}

	return nil
}

func testAccCheckContainerRegistryExists(n string, registry *containerregistry.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ContainerRegistry().Registry().Get(context.Background(), &containerregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Registry %s not found", n)
		}

		*registry = *found
		return nil
	}
}

func testAccCheckContainerRegistryName(registry *containerregistry.Registry, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if registry.Name != name {
			return fmt.Errorf("Wrong registry name: expected '%s' got '%s'", name, registry.Name)
		}
		return nil
	}
}

func testAccCheckContainerRegistryStatus(registry *containerregistry.Registry, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		registryStatus := strings.ToLower(registry.Status.String())
		if registryStatus != status {
			return fmt.Errorf("Wrong registry status: expected '%s' got '%s'", status, registryStatus)
		}
		return nil
	}
}

func testAccCheckContainerRegistryContainsLabel(registry *containerregistry.Registry, key string, value string) resource.TestCheckFunc {
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

func testAccCheckContainerRegistryDoesNotContainLabel(registry *containerregistry.Registry, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v, ok := registry.Labels[key]; ok {
			return fmt.Errorf("Expected no label for key '%s' but found one with value '%s'", key, v)
		}

		return nil
	}
}

func testAccCheckRegistyIdsEqual(registryID *string, afterUpdateRegistryID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *registryID != *afterUpdateRegistryID {
			return fmt.Errorf("Registry id has changed: before '%s', after update '%s'", *registryID, *afterUpdateRegistryID)
		}

		return nil
	}
}

func testAccContainerRegistry_update(name, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, name, folderID)
}

func testAccContainerRegistry_basic(name, folderID, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "foobar" {
  name      = "%s"
  folder_id = "%s"

  labels = {
    test_label = "%s"
  }
}
`, name, folderID, labelValue)
}
