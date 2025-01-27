package billing_cloud_binding_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	yandex_billing_cloud_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/billing_cloud_binding"
)

const billingCloudBindingBindingResource = "yandex_billing_cloud_binding.test_cloud_binding_resource_binding"
const billingCloudServiceInstanceBindingType = "cloud"

func billingInstanceTestFirstBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_1")
}
func billingInstanceTestSecondBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_2")
}

const yandexBillingServiceInstanceBindingDefaultTimeout = 1 * time.Minute

func init() {
	resource.AddTestSweepers("yandex_billing_cloud_binding", &resource.Sweeper{
		Name: "yandex_billing_cloud_binding",
		F:    testSweepBillingCloudBinding,
		Dependencies: []string{
			"yandex_resourcemanager_cloud",
		},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepBillingCloudBinding(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	if !isAllNeedleEnvsAvailable() {
		test.DebugLog("yandex_billing_cloud_binding sweeper was skeeped")
		return nil
	}

	req := &billing.ListBillableObjectBindingsRequest{
		BillingAccountId: billingInstanceTestFirstBillingAccountId(),
	}
	it := conf.SDK.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(context.Background(), req)
	result := &multierror.Error{}

	for it.Next() {
		cloudId := it.Value().BillableObject.Id
		if !sweepBillingCloudBinding(conf, cloudId) {
			result = multierror.Append(
				result,
				fmt.Errorf("failed to sweep Billing cloud binding with id %q", cloudId),
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

func isAllNeedleEnvsAvailable() bool {
	return billingInstanceTestFirstBillingAccountId() != "" && billingInstanceTestSecondBillingAccountId() != ""
}

func sweepBillingCloudBinding(conf *provider_config.Config, cloudId string) bool {
	return test.SweepWithRetry(sweepBillingCloudBindingOnce, conf, "Billing Cloud Binding", cloudId)
}

func sweepBillingCloudBindingOnce(conf *provider_config.Config, instanceId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexBillingServiceInstanceBindingDefaultTimeout)
	defer cancel()

	billableObject := &billing.BillableObject{
		Id:   instanceId,
		Type: billingCloudServiceInstanceBindingType,
	}
	req := &billing.BindBillableObjectRequest{
		BillingAccountId: billingInstanceTestSecondBillingAccountId(),
		BillableObject:   billableObject,
	}
	op, err := conf.SDK.Billing().BillingAccount().BindBillableObject(ctx, req)
	return test.HandleSweepOperation(ctx, conf, op, err)
}

func TestAccResourceBillingCloudBinding_BindExistingCloudToExistingBillingAccount(t *testing.T) {
	firstBillingAccountId := billingInstanceTestFirstBillingAccountId()
	secondBillingAccountId := billingInstanceTestSecondBillingAccountId()
	cloudId := test.GetExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(firstBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingCloudBindingExists(billingCloudBindingBindingResource),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "billing_account_id", firstBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "cloud_id", cloudId),
				),
			},
			{
				ResourceName:      billingCloudBindingBindingResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(secondBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingCloudBindingExists(billingCloudBindingBindingResource),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "billing_account_id", secondBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "cloud_id", cloudId),
				),
			},
		},
	})
}

func TestAccResourceBillingCloudBinding_BindExistingCloudToExistingBillingAccountWithSelfBind(t *testing.T) {
	firstBillingAccountId := billingInstanceTestFirstBillingAccountId()
	secondBillingAccountId := billingInstanceTestSecondBillingAccountId()
	cloudId := test.GetExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(firstBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingCloudBindingExists(billingCloudBindingBindingResource),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "billing_account_id", firstBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "cloud_id", cloudId),
				),
			},
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(firstBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingCloudBindingExists(billingCloudBindingBindingResource),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "billing_account_id", firstBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "cloud_id", cloudId),
				),
			},
			{
				ResourceName:      billingCloudBindingBindingResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(secondBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBillingCloudBindingExists(billingCloudBindingBindingResource),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "billing_account_id", secondBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingResource, "cloud_id", cloudId),
				),
			},
		},
	})
}

func TestAccResourceBillingCloudBinding_BindNonExistingCloudToExistingBillingAccount(t *testing.T) {
	nonExistingBillingAccountId := fmt.Sprintf("non-existing-billing-account-id-%s", acctest.RandString(10))
	cloudId := test.GetExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceBillingCloudBindingBindCloudToBillingAccount(nonExistingBillingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Error while requesting API binding cloud to billing account"),
			},
		},
	})
}

func TestAccResourceBillingCloudBinding_BindExistingCloudToNonExistingBillingAccount(t *testing.T) {
	billingAccountId := billingInstanceTestFirstBillingAccountId()
	cloudId := fmt.Sprintf("non-existing-cloud-id-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceBillingCloudBindingBindCloudToBillingAccount(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Error while requesting API binding cloud to billing account"),
			},
		},
	})
}

func TestAccResourceBillingCloudBinding_BindNonExistingCloudToNonExistingBillingAccount(t *testing.T) {
	billingAccountId := fmt.Sprintf("non-existing-billing-account-id-%s", acctest.RandString(10))
	cloudId := fmt.Sprintf("non-existing-cloud-id-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceBillingCloudBindingBindCloudToBillingAccount(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Error while requesting API binding cloud to billing account"),
			},
		},
	})
}

func testAccCheckBillingCloudBindingDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_billing_cloud_binding" {
			continue
		}

		id, err := yandex_billing_cloud_binding.ParseBindServiceInstanceId(rs.Primary.ID)

		if err != nil {
			return err
		}

		it := config.SDK.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
			context.Background(),
			&billing.ListBillableObjectBindingsRequest{
				BillingAccountId: billingInstanceTestFirstBillingAccountId(),
			},
		)

		for it.Next() {
			if it.Value().BillableObject.Type == billingCloudServiceInstanceBindingType &&
				it.Value().BillableObject.Id == id.ServiceInstanceId {
				return fmt.Errorf("Cloud still bound to billing account")
			}
		}
	}

	return nil
}

func testAccCheckBillingCloudBindingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		id, err := yandex_billing_cloud_binding.ParseBindServiceInstanceId(rs.Primary.ID)

		if err != nil {
			return err
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		it := config.SDK.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
			context.Background(),
			&billing.ListBillableObjectBindingsRequest{
				BillingAccountId: id.BillingAccountId,
			})

		for it.Next() {
			if it.Value().BillableObject.Type == billingCloudServiceInstanceBindingType &&
				it.Value().BillableObject.Id == id.ServiceInstanceId {
				return nil
			}
		}

		return fmt.Errorf("cloud bound to billing account not found")
	}
}

func testAccResourceBillingCloudBindingBindCloudToBillingAccount(billingAccountId string, cloudId string) string {
	return fmt.Sprintf(`
resource "yandex_billing_cloud_binding" "test_cloud_binding_resource_binding" {
	billing_account_id = "%s"
	cloud_id = "%s"
}
`, billingAccountId, cloudId)
}
