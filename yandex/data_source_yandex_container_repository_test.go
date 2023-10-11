package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

func TestAccDataSourceContainerRepository_byID(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-repository")
	var registry containerregistry.Registry

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
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
