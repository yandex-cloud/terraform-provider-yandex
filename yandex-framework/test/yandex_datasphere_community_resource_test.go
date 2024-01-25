package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider-config"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	testCommunityResourceName = "yandex_datasphere_community.test-community"
)

func init() {
	resource.AddTestSweepers("yandex_datasphere_community", &resource.Sweeper{
		Name:         "yandex_datasphere_community",
		F:            testSweepCommunity,
		Dependencies: []string{},
	})
}

func testSweepCommunity(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := conf.SDK.Datasphere().Community().CommunityIterator(
		context.Background(),
		&datasphere.ListCommunitiesRequest{OrganizationId: getExampleOrganizationID()},
	)
	result := &multierror.Error{}

	for it.Next() {
		communityId := it.Value().GetId()
		if !isTestResourseName(it.Value().GetName()) {
			continue
		}
		if !sweepCommunity(conf, communityId) {
			result = multierror.Append(
				result,
				fmt.Errorf("failed to sweep community id %q", communityId),
			)
		}
	}

	if err := it.Error(); err != nil {
		result = multierror.Append(
			result,
			fmt.Errorf("iterator error: %w", err),
		)
	}

	return result.ErrorOrNil()
}

func sweepCommunity(conf *provider_config.Config, cloudId string) bool {
	return sweepWithRetry(sweepCommunityOnce, conf, "yandex_datasphere_community", cloudId)
}

func sweepCommunityOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := conf.SDK.Datasphere().Community().Delete(ctx, &datasphere.DeleteCommunityRequest{
		CommunityId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccDatasphereCommunityResource_basic(t *testing.T) {
	var (
		communityName = testResourseName(63)

		communityDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckCommunityDestroy,
		Steps: []resource.TestStep{
			basicCommunityTestStep(communityName, communityDesc, labelKey, labelValue),
			communityImportTestStep(),
		},
	})
}

func TestAccDatasphereCommunityResource_minimalDataCreation(t *testing.T) {
	var (
		communityName = testResourseName(63)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckCommunityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCommunityConfigMinimalData(communityName),
				Check: resource.ComposeTestCheckFunc(
					testCommunityExists(testCommunityResourceName),
					resource.TestCheckResourceAttr(testCommunityResourceName, "name", communityName),
					resource.TestCheckResourceAttr(testCommunityResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testCommunityResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testCommunityResourceName, "created_by"),
					resource.TestCheckResourceAttr(testCommunityResourceName, "organization_id", getExampleOrganizationID()),
					testAccCheckCreatedAtAttr(testCommunityResourceName),
				),
			},
			communityImportTestStep(),
		},
	})
}

func TestAccDatasphereCommunityResource_update(t *testing.T) {
	var (
		communityName = testResourseName(63)
		communityDesc = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)

		communityNameUpdated = testResourseName(63)
		communityDescUpdated = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKeyUpdated      = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValueUpdated    = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		CheckDestroy:             testAccCheckCommunityDestroy,
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
		ResourceName:      testCommunityResourceName,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"billing_account_id",
		},
	}
}

func testAccCheckCommunityDestroy(s *terraform.State) error {
	config := testAccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datasphere_community" {
			continue
		}
		id := rs.Primary.ID

		_, err := config.SDK.Datasphere().Community().Get(context.Background(), &datasphere.GetCommunityRequest{
			CommunityId: id,
		})
		if err == nil {
			return fmt.Errorf("community still exists")
		}
	}

	return nil
}

func basicCommunityTestStep(communityName, communityDesc, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testCommunityBasic(communityName, communityDesc, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			testCommunityExists(testCommunityResourceName),
			resource.TestCheckResourceAttr(testCommunityResourceName, "name", communityName),
			resource.TestCheckResourceAttr(testCommunityResourceName, "description", communityDesc),
			resource.TestCheckResourceAttr(testCommunityResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testCommunityResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(testCommunityResourceName, "created_by"),
			resource.TestCheckResourceAttr(testCommunityResourceName, "organization_id", getExampleOrganizationID()),
			resource.TestCheckResourceAttr(testCommunityResourceName, "billing_account_id", getBillingAccountId()),
			testAccCheckCreatedAtAttr(testCommunityResourceName),
		),
	}
}

func testCommunityExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.(*yandex_framework.Provider).GetConfig()
		a := s.RootModule().Resources
		fmt.Printf("%s", a)
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		found, err := config.SDK.Datasphere().Community().Get(context.Background(), &datasphere.GetCommunityRequest{
			CommunityId: id,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("community not found")
		}

		return nil
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
`, name, desc, getBillingAccountId(), labelKey, labelValue, getExampleOrganizationID())
}

func testCommunityConfigMinimalData(name string) string {
	return fmt.Sprintf(`
resource "yandex_datasphere_community" "test-community" {
  name = "%s"
  organization_id = "%s"
}
`, name, getExampleOrganizationID())
}
