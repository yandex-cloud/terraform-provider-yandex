package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccDataSourceContainerRegistry_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckContainerRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRegistryConfig(registryName, folderID, label, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_container_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"name", registryName),
					resource.TestCheckResourceAttrSet("data.yandex_container_registry.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_container_registry.source"),
				),
			},
		},
	})
}

func TestAccDataSourceContainerRegistry_byName(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	label := acctest.RandomWithPrefix("label-value")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy:             testAccCheckContainerRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRegistryConfig(registryName, folderID, label, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField("data.yandex_container_registry.source", "registry_id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"name", registryName),
					resource.TestCheckResourceAttrSet("data.yandex_container_registry.source",
						"id"),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_container_registry.source",
						"labels.test_label", label),
					testAccCheckCreatedAtAttr("data.yandex_container_registry.source"),
				),
			},
		},
	})
}

func testAccCheckContainerRegistryDestroy(s *terraform.State) error {
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
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Container Registry still exists")
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
			return fmt.Errorf("Container Registry %s not found", n)
		}

		*registry = *found
		return nil
	}
}

func testAccDataSourceContainerRegistryConfig(folderID, name, labelValue string, useID bool) string {
	if useID {
		return testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue) + containerRegistryDataByIDConfig
	}

	return testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue) + containerRegistryDataByNameConfig
}

func testAccDataSourceContainerRegistryResourceConfig(folderID, name, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "foobar" {
  name     = "%s"
  folder_id = "%s"

  labels = {
    test_label = "%s"
  }
}
`, folderID, name, labelValue)
}

const containerRegistryDataByIDConfig = `
data "yandex_container_registry" "source" {
  registry_id = "${yandex_container_registry.foobar.id}"
}
`

const containerRegistryDataByNameConfig = `
data "yandex_container_registry" "source" {
  name = "${yandex_container_registry.foobar.name}"
}
`
