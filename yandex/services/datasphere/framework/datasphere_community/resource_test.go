package datasphere_community_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/testfixtures"
)

func TestAccDatasphereCommunityResource_basic(t *testing.T) {
	var (
		communityName = test.ResourceName(63)

		communityDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckCommunityDestroy(communityClient),
		Steps: []resource.TestStep{
			basicCommunityTestStep(communityName, communityDesc, labelKey, labelValue),
			communityImportTestStep(),
		},
	})
}

func TestAccDatasphereCommunityResource_minimalDataCreation(t *testing.T) {
	var (
		communityName = test.ResourceName(63)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckCommunityDestroy(communityClient),
		Steps: []resource.TestStep{
			{
				Config: testCommunityConfigMinimalData(communityName),
				Check: resource.ComposeTestCheckFunc(
					fixtures.CommunityExists(communityClient, fixtures.CommunityResourceName),
					resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "name", communityName),
					resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(fixtures.CommunityResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(fixtures.CommunityResourceName, "created_by"),
					resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "organization_id", test.GetExampleOrganizationID()),
					test.AccCheckCreatedAtAttr(fixtures.CommunityResourceName),
				),
			},
			communityImportTestStep(),
		},
	})
}

func TestAccDatasphereCommunityResource_update(t *testing.T) {
	var (
		communityName = test.ResourceName(63)
		communityDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

		communityNameUpdated = test.ResourceName(63)
		communityDescUpdated = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKeyUpdated      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValueUpdated    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             fixtures.AccCheckCommunityDestroy(communityClient),
		Steps: []resource.TestStep{
			basicCommunityTestStep(communityName, communityDesc, labelKey, labelValue),
			communityImportTestStep(),
			basicCommunityTestStep(communityNameUpdated, communityDescUpdated, labelKeyUpdated, labelValueUpdated),
			communityImportTestStep(),
		},
	})
}

func communityImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      fixtures.CommunityResourceName,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"billing_account_id",
		},
	}
}

func basicCommunityTestStep(communityName, communityDesc, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testCommunityBasic(communityName, communityDesc, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			fixtures.CommunityExists(communityClient, fixtures.CommunityResourceName),
			resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "name", communityName),
			resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "description", communityDesc),
			resource.TestCheckResourceAttr(fixtures.CommunityResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(fixtures.CommunityResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(fixtures.CommunityResourceName, "created_by"),
			resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "organization_id", test.GetExampleOrganizationID()),
			resource.TestCheckResourceAttr(fixtures.CommunityResourceName, "billing_account_id", test.GetBillingAccountId()),
			test.AccCheckCreatedAtAttr(fixtures.CommunityResourceName),
		),
	}
}

func testCommunityBasic(name string, desc string, labelKey string, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  description = "%s"
  billing_account_id = "%s"
  labels = {
    "%s": "%s"
  }
  organization_id = "%s"
}
`, name, desc, test.GetBillingAccountId(), labelKey, labelValue, test.GetExampleOrganizationID())
}

func testCommunityConfigMinimalData(name string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  organization_id = "%s"
}
`, name, test.GetExampleOrganizationID())
}
