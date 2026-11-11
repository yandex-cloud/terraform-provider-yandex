package acceptance

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
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_container_registry "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/containerregistry/pluginsdk"
)

const containerRegistryResource = "yandex_container_registry.test-registry"

func importContainerRegistryIDFunc(registry *containerregistry.Registry, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return registry.Id + "," + role, nil
	}
}

func TestAccContainerRegistryIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var registry containerregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-container-registry")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRegistryIamBindingBasic(registryName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckContainerRegistryExists(containerRegistryResource, &registry),
					suite.testAccCheckContainerRegistryIam(containerRegistryResource, role, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_container_registry_iam_binding.puller",
				ImportStateIdFunc:                    importContainerRegistryIDFunc(&registry, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

func TestAccContainerRegistryIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var registry containerregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-container-registry")

	role := "container-registry.images.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccContainerRegistry(registryName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckContainerRegistryExists(containerRegistryResource, &registry),
					suite.testAccCheckContainerRegistryEmptyIam(containerRegistryResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccContainerRegistryIamBindingBasic(registryName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckContainerRegistryIam(containerRegistryResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccContainerRegistry(registryName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckContainerRegistryEmptyIam(containerRegistryResource),
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

func (suite testSuite) testAccCheckContainerRegistryEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getContainerRegistryResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckContainerRegistryIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getContainerRegistryResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getContainerRegistryResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_container_registry.GetContainerRegistryAccessBindings(config.Context(), config, rs.Primary.ID)
}
