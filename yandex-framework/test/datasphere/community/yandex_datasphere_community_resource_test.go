package community

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test"
	dataspheretest "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/test/datasphere"
)

func init() {
	resource.AddTestSweepers("yandex_datasphere_community", &resource.Sweeper{
		Name:         "yandex_datasphere_community",
		F:            testSweepCommunity,
		Dependencies: []string{},
	})
}

func testSweepCommunity(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := conf.SDK.Datasphere().Community().CommunityIterator(
		context.Background(),
		&datasphere.ListCommunitiesRequest{OrganizationId: test.GetExampleOrganizationID()},
	)
	result := &multierror.Error{}

	for it.Next() {
		communityId := it.Value().GetId()
		if !test.IsTestResourceName(it.Value().GetName()) {
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
	return test.SweepWithRetry(sweepCommunityOnce, conf, "yandex_datasphere_community", cloudId)
}

func sweepCommunityOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := conf.SDK.Datasphere().Community().Delete(ctx, &datasphere.DeleteCommunityRequest{
		CommunityId: id,
	})
	return test.HandleSweepOperation(ctx, conf, op, err)
}

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
		CheckDestroy:             dataspheretest.AccCheckCommunityDestroy,
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
		CheckDestroy:             dataspheretest.AccCheckCommunityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCommunityConfigMinimalData(communityName),
				Check: resource.ComposeTestCheckFunc(
					dataspheretest.CommunityExists(dataspheretest.CommunityResourceName),
					resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "name", communityName),
					resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(dataspheretest.CommunityResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(dataspheretest.CommunityResourceName, "created_by"),
					resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "organization_id", test.GetExampleOrganizationID()),
					test.AccCheckCreatedAtAttr(dataspheretest.CommunityResourceName),
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
		CheckDestroy:             dataspheretest.AccCheckCommunityDestroy,
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
		ResourceName:      dataspheretest.CommunityResourceName,
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
			dataspheretest.CommunityExists(dataspheretest.CommunityResourceName),
			resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "name", communityName),
			resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "description", communityDesc),
			resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(dataspheretest.CommunityResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(dataspheretest.CommunityResourceName, "created_by"),
			resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "organization_id", test.GetExampleOrganizationID()),
			resource.TestCheckResourceAttr(dataspheretest.CommunityResourceName, "billing_account_id", test.GetBillingAccountId()),
			test.AccCheckCreatedAtAttr(dataspheretest.CommunityResourceName),
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
