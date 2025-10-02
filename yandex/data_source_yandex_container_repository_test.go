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

func TestAccDataSourceContainerRepository_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-repository")
	var registry containerregistry.Registry

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRepositoryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRepositoryConfig(registryName, repositoryNameSuffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.my-reg", &registry),
					testAccCheckResourceIDField("data.yandex_container_repository.source", "repository_id"),
					testAccCheckDataContainerRepositoryName(&registry, repositoryNameSuffix),
					resource.TestCheckResourceAttrSet("data.yandex_container_repository.source", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceContainerRepository_byName(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-repository")
	var registry containerregistry.Registry

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRepositoryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceContainerRepositoryConfig(registryName, repositoryNameSuffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.my-reg", &registry),
					testAccCheckResourceIDField("data.yandex_container_repository.source", "repository_id"),
					testAccCheckDataContainerRepositoryName(&registry, repositoryNameSuffix),
					resource.TestCheckResourceAttrSet("data.yandex_container_repository.source", "id"),
				),
			},
		},
	})
}

func testAccCheckContainerRepositoryDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_repository" {
			continue
		}

		_, err := config.sdk.ContainerRegistry().Repository().Get(context.Background(), &containerregistry.GetRepositoryRequest{
			RepositoryId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Container Repository still exists")
		}
	}

	return nil
}

func testAccCheckContainerRepositoryExists(n string, repository *containerregistry.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ContainerRegistry().Repository().Get(context.Background(), &containerregistry.GetRepositoryRequest{
			RepositoryId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Container Repository %s not found", n)
		}

		*repository = *found
		return nil
	}
}

func testAccDataSourceContainerRepositoryConfig(registryName, repositoryNameSuffix string, useID bool) string {
	if useID {
		return testAccDataSourceContainerRepositoryResourceConfig(registryName, repositoryNameSuffix) + containerRepositoryDataByIDConfig
	}

	return testAccDataSourceContainerRepositoryResourceConfig(registryName, repositoryNameSuffix) + containerRepositoryDataByNameConfig
}

func testAccCheckDataContainerRepositoryName(registry *containerregistry.Registry, repositoryNameSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := repositoryFullName(registry.Id, repositoryNameSuffix)
		return resource.TestCheckResourceAttr("data.yandex_container_repository.source", "name", name)(s)
	}
}

func repositoryFullName(registryId, repositoryNameSuffix string) string {
	return fmt.Sprintf("%s/%s", registryId, repositoryNameSuffix)
}

func testAccDataSourceContainerRepositoryResourceConfig(registryName, repositoryNameSuffix string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "my-reg" {
  name = "%s"
}

resource "yandex_container_repository" "my-repo" {
	name = "${yandex_container_registry.my-reg.id}/%s"
}
`, registryName, repositoryNameSuffix)
}

const containerRepositoryDataByIDConfig = `
data "yandex_container_repository" "source" {
  repository_id = "${yandex_container_repository.my-repo.id}"
}
`

const containerRepositoryDataByNameConfig = `
data "yandex_container_repository" "source" {
  name = "${yandex_container_repository.my-repo.name}"
}
`
