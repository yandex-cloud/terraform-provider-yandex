package yandex_container_repository_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	containerregistrysdk "github.com/yandex-cloud/go-sdk/services/containerregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

//revive:disable:var-naming
func TestAccContainerRepository_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-repository")
	var registry containerregistry.Registry
	var repository containerregistry.Repository

	resource.Test(t, resource.TestCase{
		PreCheck: func() { test.AccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRepositoryDestroy,
		),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccContainerRepository_basic(registryName, repositoryNameSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.my-reg", &registry),
					testAccCheckContainerRepositoryExists("yandex_container_repository.my-repo", &repository),
					testAccCheckContainerRepositoryName(&registry, &repository, repositoryNameSuffix),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccContainerRepository_basic(registryName, repositoryNameSuffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

//revive:disable:var-naming
func TestAccContainerRepository_basic(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-repository")
	var registry containerregistry.Registry
	var repository containerregistry.Repository

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRepositoryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRepository_basic(registryName, repositoryNameSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.my-reg", &registry),
					testAccCheckContainerRepositoryExists("yandex_container_repository.my-repo", &repository),
					testAccCheckContainerRepositoryName(&registry, &repository, repositoryNameSuffix),
				),
			},
		},
	})
}

func TestAccContainerRepository_children(t *testing.T) {
	t.Parallel()

	registryName := acctest.RandomWithPrefix("tf-registry")
	child1NameSuffix := acctest.RandomWithPrefix("tf-repository") + "/level-1"
	child2NameSuffix := child1NameSuffix + "/level-2"

	var registry containerregistry.Registry
	var child1Repository containerregistry.Repository
	var child2Repository containerregistry.Repository

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckContainerRegistryDestroy,
			testAccCheckContainerRepositoryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRepository_children(registryName, child1NameSuffix, child2NameSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists("yandex_container_registry.my-reg", &registry),
					testAccCheckContainerRepositoryExists("yandex_container_repository.child-1", &child1Repository),
					testAccCheckContainerRepositoryName(&registry, &child1Repository, child1NameSuffix),
					testAccCheckContainerRepositoryExists("yandex_container_repository.child-2", &child2Repository),
					testAccCheckContainerRepositoryName(&registry, &child2Repository, child2NameSuffix),
				),
			},
		},
	})
}

func testAccCheckContainerRepositoryDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_repository" {
			continue
		}

		_, err := containerregistrysdk.NewRepositoryClient(config.SDKv2).Get(context.Background(), &containerregistry.GetRepositoryRequest{
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

func testAccCheckContainerRegistryDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_container_registry" {
			continue
		}

		_, err := containerregistrysdk.NewRegistryClient(config.SDKv2).Get(context.Background(), &containerregistry.GetRegistryRequest{
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

func testAccCheckContainerRepositoryExists(n string, repository *containerregistry.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := containerregistrysdk.NewRepositoryClient(config.SDKv2).Get(context.Background(), &containerregistry.GetRepositoryRequest{
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

func testAccCheckContainerRegistryExists(n string, registry *containerregistry.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := containerregistrysdk.NewRegistryClient(config.SDKv2).Get(context.Background(), &containerregistry.GetRegistryRequest{
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

func testAccCheckContainerRepositoryName(registry *containerregistry.Registry, repository *containerregistry.Repository, repositoryNameSuffix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := repositoryFullName(registry.Id, repositoryNameSuffix)
		if repository.Name != name {
			return fmt.Errorf("Wrong Container Repository name: expected '%s' got '%s'", name, repository.Name)
		}
		return nil
	}
}

func testAccContainerRepository_basic(registryName, repositoryNameSuffix string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "my-reg" {
  name = "%s"
}

resource "yandex_container_repository" "my-repo" {
  name = "${yandex_container_registry.my-reg.id}/%s"
}
`, registryName, repositoryNameSuffix)
}

func testAccContainerRepository_children(registryName, child1NameSuffix, child2NameSuffix string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "my-reg" {
  name = "%s"
}

resource "yandex_container_repository" "child-2" {
  name = "${yandex_container_registry.my-reg.id}/%s"
}

resource "yandex_container_repository" "child-1" {
  name = "${yandex_container_registry.my-reg.id}/%s"
}
`, registryName, child2NameSuffix, child1NameSuffix)
}

func repositoryFullName(registryId, repositoryNameSuffix string) string {
	return fmt.Sprintf("%s/%s", registryId, repositoryNameSuffix)
}
