package cloudregistry_folder_iam_member

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	cloudregistryv1sdk "github.com/yandex-cloud/go-sdk/services/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	cloudRegistryFolderResource  = "yandex_cloudregistry_folder.test-folder"
	cloudRegistryArtifactIamRole = "cloud-registry.artifacts.puller"
	defaultListPageSize          = 1000
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRegistryFolderIamMember_lifecycle(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")
	firstMember := "system:allUsers"
	secondMember := "system:allAuthenticatedUsers"
	var artifactID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryFolderIamMemberConfig(registryName, cloudRegistryArtifactIamRole, firstMember, secondMember),
				Check: resource.ComposeTestCheckFunc(
					captureCloudRegistryArtifactID(cloudRegistryFolderResource, &artifactID),
					testAccCheckCloudRegistryArtifactIam(cloudRegistryFolderResource, cloudRegistryArtifactIamRole, []string{firstMember, secondMember}),
				),
			},
			{
				PreConfig: func() {
					removeCloudRegistryFolderIamMember(t, artifactID, cloudRegistryArtifactIamRole, firstMember)
				},
				Config:             testAccCloudRegistryFolderIamMemberConfig(registryName, cloudRegistryArtifactIamRole, firstMember, secondMember),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCloudRegistryFolderIamMemberConfig(registryName, cloudRegistryArtifactIamRole, firstMember, secondMember),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryArtifactIam(cloudRegistryFolderResource, cloudRegistryArtifactIamRole, []string{firstMember, secondMember}),
				),
			},
			{
				Config: testAccCloudRegistryFolderIamMemberConfig(registryName, cloudRegistryArtifactIamRole, secondMember),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryArtifactIam(cloudRegistryFolderResource, cloudRegistryArtifactIamRole, []string{secondMember}),
				),
			},
			{
				ResourceName:                         "yandex_cloudregistry_folder_iam_member.puller_0",
				ImportStateIdFunc:                    importCloudRegistryFolderIamMemberIDFunc(cloudRegistryFolderResource, cloudRegistryArtifactIamRole, secondMember),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "artifact_id",
				ImportStateVerifyIgnore:              []string{"sleep_after"},
			},
			{
				Config: testAccCloudRegistryFolder(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryArtifactEmptyIam(cloudRegistryFolderResource),
				),
			},
		},
	})
}

func importCloudRegistryFolderIamMemberIDFunc(artifactResource, role, member string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[artifactResource]
		if !ok {
			return "", fmt.Errorf("can't find %s in state", artifactResource)
		}
		return rs.Primary.ID + "," + role + "," + member, nil
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

func testAccCloudRegistryFolderIamMemberConfig(registryName, role string, members ...string) string {
	config := testAccCloudRegistryFolder(registryName)
	for i, member := range members {
		config += fmt.Sprintf(`
resource "yandex_cloudregistry_folder_iam_member" "puller_%d" {
  artifact_id = yandex_cloudregistry_folder.test-folder.id
  role        = "%s"
  member      = "%s"
  sleep_after = 30
}
`, i, role, member)
	}
	return config
}

func testAccCheckCloudRegistryArtifactEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryArtifactAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckCloudRegistryArtifactIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getCloudRegistryArtifactAccessBindings(s, resourceName)
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

func getCloudRegistryArtifactAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
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
			return nil, fmt.Errorf("Error retrieving access bindings of Cloud Registry artifact %s: %w", rs.Primary.ID, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return bindings, nil
}

func captureCloudRegistryArtifactID(resourceName string, artifactID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}
		*artifactID = rs.Primary.ID
		return nil
	}
}

func removeCloudRegistryFolderIamMember(t *testing.T, artifactID, role, member string) {
	t.Helper()

	memberParts := strings.SplitN(member, ":", 2)
	if len(memberParts) != 2 {
		t.Fatalf("invalid member %q", member)
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	op, err := cloudregistryv1sdk.NewArtifactClient(config.SDKv2).UpdateAccessBindings(ctx, &access.UpdateAccessBindingsRequest{
		ResourceId: artifactID,
		AccessBindingDeltas: []*access.AccessBindingDelta{
			{
				Action: access.AccessBindingAction_REMOVE,
				AccessBinding: &access.AccessBinding{
					RoleId: role,
					Subject: &access.Subject{
						Type: memberParts[0],
						Id:   memberParts[1],
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("remove IAM member: %v", err)
	}

	if _, err = op.Wait(ctx); err != nil {
		t.Fatalf("wait for removing IAM member: %v", err)
	}
}
