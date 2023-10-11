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

const containerRegistryResource = "yandex_container_registry.test-registry"

func importContainerRegistryIDFunc(registry *containerregistry.Registry, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return registry.Id + " " + role, nil
	}
}

func TestAccContainerRegistryIamBinding_basic(t *testing.T) {
	var registry containerregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-container-registry")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistryIamBindingBasic(registryName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists(containerRegistryResource, &registry),
					testAccCheckContainerRegistryIam(containerRegistryResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_container_registry_iam_binding.puller",
				ImportStateIdFunc: importContainerRegistryIDFunc(&registry, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccContainerRegistryIamBinding_remove(t *testing.T) {
	var registry containerregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-container-registry")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccContainerRegistry(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryExists(containerRegistryResource, &registry),
					testAccCheckContainerRegistryEmptyIam(containerRegistryResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccContainerRegistryIamBindingBasic(registryName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryIam(containerRegistryResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccContainerRegistry(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRegistryEmptyIam(containerRegistryResource),
				),
			},
		},
	})
}

func testAccContainerRegistryIamBindingBasic(registryName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "test-registry" {
  name       = "%s"
}

resource "yandex_container_registry_iam_binding" "puller" {
  registry_id = yandex_container_registry.test-registry.id
  role        = "%s"
  members     = ["%s"]
}
`, registryName, role, userID)
}

func testAccContainerRegistry(registryName string) string {
	return fmt.Sprintf(`
resource "yandex_container_registry" "test-registry" {
  name       = "%s"
}
`, registryName)
}

func testAccCheckContainerRegistryEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getContainerRegistryResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckContainerRegistryIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getContainerRegistryResourceAccessBindings(s, resourceName)
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

func getContainerRegistryResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getContainerRegistryAccessBindings(config.Context(), config, rs.Primary.ID)
}
