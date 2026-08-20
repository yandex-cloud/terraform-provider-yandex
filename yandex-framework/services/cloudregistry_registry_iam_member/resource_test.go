package cloudregistry_registry_iam_member

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	cloudregistryv1sdk "github.com/yandex-cloud/go-sdk/services/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	cloudRegistryResource = "yandex_cloudregistry_registry.test-registry"
	cloudRegistryIamRole  = "cloud-registry.artifacts.puller"
	defaultListSize       = 1000
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func importCloudRegistryIDFunc(registry *cloudregistry.Registry, role, member string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return registry.Id + "," + role + "," + member, nil
	}
}

func TestAccCloudRegistryIamMember_lifecycle(t *testing.T) {
	var registry cloudregistry.Registry
	registryName := acctest.RandomWithPrefix("tf-cloud-registry")
	firstMember := "system:allUsers"
	secondMember := "system:allAuthenticatedUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryIamMemberConfig(registryName, cloudRegistryIamRole, firstMember, secondMember),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource, &registry),
					testAccCheckCloudRegistryIam(cloudRegistryResource, cloudRegistryIamRole, []string{firstMember, secondMember}),
				),
			},
			{
				PreConfig: func() {
					removeCloudRegistryRegistryIamMember(t, registry.Id, cloudRegistryIamRole, firstMember)
				},
				Config:             testAccCloudRegistryIamMemberConfig(registryName, cloudRegistryIamRole, firstMember, secondMember),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCloudRegistryIamMemberConfig(registryName, cloudRegistryIamRole, firstMember, secondMember),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryIam(cloudRegistryResource, cloudRegistryIamRole, []string{firstMember, secondMember}),
				),
			},
			{
				Config: testAccCloudRegistryIamMemberConfig(registryName, cloudRegistryIamRole, secondMember),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryIam(cloudRegistryResource, cloudRegistryIamRole, []string{secondMember}),
				),
			},
			{
				ResourceName:                         "yandex_cloudregistry_registry_iam_member.puller_0",
				ImportStateIdFunc:                    importCloudRegistryIDFunc(&registry, cloudRegistryIamRole, secondMember),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
				ImportStateVerifyIgnore:              []string{"sleep_after"},
			},
			{
				Config: testAccCloudRegistry(registryName, "DOCKER", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryEmptyIam(cloudRegistryResource),
				),
			},
		},
	})
}

func testAccCloudRegistryIamMemberConfig(registryName, role string, members ...string) string {
	config := testAccCloudRegistry(registryName, "DOCKER", "LOCAL")
	for i, member := range members {
		config += fmt.Sprintf(`
resource "yandex_cloudregistry_registry_iam_member" "puller_%d" {
  registry_id = yandex_cloudregistry_registry.test-registry.id
  role        = "%s"
  member      = "%s"
  sleep_after = 30
}
`, i, role, member)
	}
	return config
}

func testAccCloudRegistry(registryName, kind, typeName string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "test-registry" {
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

		found, err := cloudregistryv1sdk.NewRegistryClient(config.SDKv2).Get(context.Background(), &cloudregistry.GetRegistryRequest{
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
		resp, err := cloudregistryv1sdk.NewRegistryClient(config.SDKv2).ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
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

func removeCloudRegistryRegistryIamMember(t *testing.T, registryID, role, member string) {
	t.Helper()

	memberParts := strings.SplitN(member, ":", 2)
	if len(memberParts) != 2 {
		t.Fatalf("invalid member %q", member)
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	op, err := cloudregistryv1sdk.NewRegistryClient(config.SDKv2).UpdateAccessBindings(ctx, &access.UpdateAccessBindingsRequest{
		ResourceId: registryID,
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
