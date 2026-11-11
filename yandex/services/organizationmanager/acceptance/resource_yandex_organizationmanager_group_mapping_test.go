package acceptance

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	sdkTerraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

func TestAccOrganizationManagerGroupMapping(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	info := newSamlFederationInfo()
	federationName := info.getResourceName(true)
	resourceName := "yandex_organizationmanager_group_mapping.acceptance"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccOrganizationManagerGroupMappingCheckDestroy(federationName),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerGroupMapping(federationName, "true"),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationManagerGroupMappingExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerGroupMapping(federationName, "false"),
				Check: suite.testAccCheckOrganizationManagerGroupMappingExists(federationName),
			},
		},
	})
}

func testAccOrganizationManagerGroupMapping(federationName, enabled string) string {
	return fmt.Sprintf(`resource yandex_organizationmanager_group_mapping "acceptance" {
  federation_id = %s.id
  enabled = %s
}
`, federationName, enabled)
}

func (suite testSuite) testAccCheckOrganizationManagerGroupMappingExists(federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := organizationmanagersdk.NewGroupMappingClient(config.SDK)

		federationId := s.RootModule().Resources[federationName].Primary.ID

		_, err := client.Get(context.Background(), &organizationmanager.GetGroupMappingRequest{
			FederationId: federationId,
		})

		return err
	}
}

func (suite testSuite) testAccOrganizationManagerGroupMappingCheckDestroy(federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := organizationmanagersdk.NewGroupMappingClient(config.SDK)

		federationId := s.RootModule().Resources[federationName].Primary.ID

		_, err := client.Get(context.Background(), &organizationmanager.GetGroupMappingRequest{
			FederationId: federationId,
		})

		if err == nil {
			return fmt.Errorf("group mapping still exists")
		}

		return nil
	}
}

func TestOrganizationManagerGroupMappingEnabledDiff(t *testing.T) {
	for _, tc := range []struct {
		name       string
		prior      bool
		configured interface{}
		wantChange bool
	}{
		{name: "omitted preserves enabled mapping", prior: true},
		{name: "omitted preserves disabled mapping", prior: false},
		{name: "explicit false disables mapping", prior: true, configured: false, wantChange: true},
		{name: "explicit true enables mapping", prior: false, configured: true, wantChange: true},
		{name: "explicit true preserves enabled mapping", prior: true, configured: true},
		{name: "explicit false preserves disabled mapping", prior: false, configured: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := svc_organization_manager.ResourceYandexOrganizationManagerGroupMapping()
			// State after importing and refreshing an existing mapping.
			state := &sdkTerraform.InstanceState{
				ID: "test-federation",
				Attributes: map[string]string{
					"federation_id": "test-federation",
					"enabled":       strconv.FormatBool(tc.prior),
				},
			}
			config := map[string]interface{}{"federation_id": state.ID}
			if tc.configured != nil {
				config["enabled"] = tc.configured
			}

			diff, err := r.Diff(context.Background(), state, sdkTerraform.NewResourceConfigRaw(config), nil)
			require.NoError(t, err)
			if !tc.wantChange {
				require.True(t, diff == nil || diff.Empty(), "unexpected diff: %#v", diff)
				return
			}
			require.NotNil(t, diff)
			change := diff.Attributes["enabled"]
			require.NotNil(t, change)
			require.Equal(t, strconv.FormatBool(tc.prior), change.Old)
			require.Equal(t, strconv.FormatBool(tc.configured.(bool)), change.New)
			require.False(t, change.NewComputed)
			require.False(t, diff.RequiresNew(), "changing enabled must update the existing mapping")
		})
	}
}
