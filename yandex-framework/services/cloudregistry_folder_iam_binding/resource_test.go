package cloudregistry_folder_iam_binding

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	cloudregistryv1sdk "github.com/yandex-cloud/go-sdk/services/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	cloudRegistryFolderResource = "yandex_cloudregistry_folder.test-folder"
	cloudRegistryFolderIamRole  = "cloud-registry.artifacts.puller"
	defaultListPageSize         = 1000
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRegistryFolderIamBinding_basic(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryFolderIamBindingBasic(registryName, cloudRegistryFolderIamRole, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryFolderIam(cloudRegistryFolderResource, cloudRegistryFolderIamRole, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_cloudregistry_folder_iam_binding.puller",
				ImportStateIdFunc:                    importCloudRegistryFolderIamBindingIDFunc(cloudRegistryFolderResource, cloudRegistryFolderIamRole),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "artifact_id",
				ImportStateVerifyIgnore:              []string{"sleep_after"},
			},
		},
	})
}

func TestAccCloudRegistryFolderIamBinding_remove(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			// Prepare a folder without any bindings.
			{
				Config: testAccCloudRegistryFolder(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryFolderEmptyIam(cloudRegistryFolderResource),
				),
			},
			// Apply IAM binding.
			{
				Config: testAccCloudRegistryFolderIamBindingBasic(registryName, cloudRegistryFolderIamRole, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryFolderIam(cloudRegistryFolderResource, cloudRegistryFolderIamRole, []string{userID}),
				),
			},
			// Remove the binding.
			{
				Config: testAccCloudRegistryFolder(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryFolderEmptyIam(cloudRegistryFolderResource),
				),
			},
		},
	})
}

func importCloudRegistryFolderIamBindingIDFunc(folderResource, role string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[folderResource]
		if !ok {
			return "", fmt.Errorf("can't find %s in state", folderResource)
		}
		return rs.Primary.ID + "," + role, nil
	}
}

func testAccCloudRegistryFolder(registryName string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "test-registry" {
  name = "%s"
  kind = "DOCKER"
  type = "LOCAL"
}

resource "yandex_cloudregistry_folder" "test-folder" {
  registry_id = yandex_cloudregistry_registry.test-registry.id
  path        = "common-artifacts/some-folder"
}
`, registryName)
}

func testAccCloudRegistryFolderIamBindingBasic(registryName, role, userID string) string {
	return testAccCloudRegistryFolder(registryName) + fmt.Sprintf(`
resource "yandex_cloudregistry_folder_iam_binding" "puller" {
  artifact_id = yandex_cloudregistry_folder.test-folder.id
  role        = "%s"
  members     = ["%s"]
  sleep_after = 30
}
`, role, userID)
}

func testAccCheckCloudRegistryFolderEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryFolderAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckCloudRegistryFolderIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryFolderAccessBindings(s, resourceName)
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

func getCloudRegistryFolderAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	bindings := []*access.AccessBinding{}
	pageToken := ""

	for {
		resp, err := cloudregistryv1sdk.NewArtifactClient(config.SDKv2).ListAccessBindings(context.Background(), &access.ListAccessBindingsRequest{
			ResourceId: rs.Primary.ID,
			PageSize:   defaultListPageSize,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error retrieving access bindings of Cloud Registry folder %s: %w", rs.Primary.ID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return bindings, nil
}
