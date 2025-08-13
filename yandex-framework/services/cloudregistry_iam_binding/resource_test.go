package cloudregistry_iam_binding_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const cloudRegistryResource = "yandex_cloud_registry.test-registry"
const defaultListSize = 1000

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func importCloudRegistryIDFunc(registry *cloudregistry.Registry, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return registry.Id + " " + role, nil
	}
}

func TestAccCloudRegistryIamBinding_basic(t *testing.T) {
	var registry cloudregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-cloud-registry")

	role := "cloud-registry.artifacts.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryIamBindingBasic(registryName, "DOCKER", "LOCAL", role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource, &registry),
					testAccCheckCloudRegistryIam(cloudRegistryResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_cloud_registry_iam_binding.puller",
				ImportStateIdFunc: importCloudRegistryIDFunc(&registry, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistryIamBinding_remove(t *testing.T) {
	var registry cloudregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-cloud-registry")

	role := "cloud-registry.artifacts.puller"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccCloudRegistry(registryName, "DOCKER", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource, &registry),
					testAccCheckCloudRegistryEmptyIam(cloudRegistryResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccCloudRegistryIamBindingBasic(registryName, "DOCKER", "LOCAL", role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryIam(cloudRegistryResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccCloudRegistry(registryName, "DOCKER", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryEmptyIam(cloudRegistryResource),
				),
			},
		},
	})
}

func testAccCloudRegistryIamBindingBasic(registryName, kind, typeName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_cloud_registry" "test-registry" {
  name       = "%s"
  kind       = "%s"
  type		 = "%s"
}

resource "yandex_cloud_registry_iam_binding" "puller" {
  registry_id = yandex_cloud_registry.test-registry.id
  role        = "%s"
  members     = ["%s"]
}
`, registryName, kind, typeName, role, userID)
}

func testAccCloudRegistry(registryName, kind, typeName string) string {
	return fmt.Sprintf(`
resource "yandex_cloud_registry" "test-registry" {
  name       = "%s"
  kind       = "%s"
  type		 = "%s"
}
`, registryName, kind, typeName)
}

func testAccCheckCloudRegistryEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckCloudRegistryIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryResourceAccessBindings(s, resourceName)
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

func getCloudRegistryResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getCloudRegistryAccessBindings(context.Background(), config, rs.Primary.ID)
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

func getCloudRegistryAccessBindings(ctx context.Context, config provider_config.Config, registryID string) ([]*access.AccessBinding, error) {
	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := config.SDK.CloudRegistry().Registry().ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: registryID,
			PageSize:   defaultListSize,
			PageToken:  pageToken,
		})

		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of Cloud Registry %s: %w", registryID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
