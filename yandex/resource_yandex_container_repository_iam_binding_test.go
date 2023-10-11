package yandex

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

const containerRepositoryResource = "yandex_container_repository.test-repository"

func importContainerRepositoryIDFunc(repository *containerregistry.Repository, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return repository.Id + " " + role, nil
	}
}

func TestAccContainerRepositoryIamBinding_basic(t *testing.T) {
	var repository containerregistry.Repository
	registryName := acctest.RandomWithPrefix("tf-container-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-container-repository")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRepositoryIamBindingBasic(registryName, repositoryNameSuffix, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRepositoryExists(containerRepositoryResource, &repository),
					testAccCheckContainerRepositoryIam(containerRepositoryResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_container_repository_iam_binding.puller",
				ImportStateIdFunc: importContainerRepositoryIDFunc(&repository, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccContainerRepositoryIamBinding_remove(t *testing.T) {
	var repository containerregistry.Repository
	registryName := acctest.RandomWithPrefix("tf-container-registry")
	repositoryNameSuffix := acctest.RandomWithPrefix("tf-container-repository")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccContainerRepository(registryName, repositoryNameSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRepositoryExists(containerRepositoryResource, &repository),
					testAccCheckContainerRepositoryEmptyIam(containerRepositoryResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccContainerRepositoryIamBindingBasic(registryName, repositoryNameSuffix, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRepositoryIam(containerRepositoryResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccContainerRepository(registryName, repositoryNameSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRepositoryEmptyIam(containerRepositoryResource),
				),
			},
		},
	})
}

func testAccContainerRepositoryIamBindingBasic(registryName, repositoryNameSuffix, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "test-registry" {
  name       = "%s"
}

resource "yandex_container_repository" "test-repository" {
  name       = "${yandex_container_registry.test-registry.id}/%s"
}

resource "yandex_container_repository_iam_binding" "puller" {
  repository_id = yandex_container_repository.test-repository.id
  role        = "%s"
  members     = ["%s"]
}
`, registryName, repositoryNameSuffix, role, userID)
}

func testAccContainerRepository(registryName, repositoryNameSuffix string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "test-registry" {
  name       = "%s"
}

resource "yandex_container_repository" "test-repository" {
  name       = "${yandex_container_registry.test-registry.id}/%s"
}
`, registryName, repositoryNameSuffix)
}

func testAccCheckContainerRepositoryEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getContainerRepositoryResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckContainerRepositoryIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getContainerRepositoryResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func getContainerRepositoryResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getContainerRepositoryAccessBindings(config.Context(), config, rs.Primary.ID)
}
