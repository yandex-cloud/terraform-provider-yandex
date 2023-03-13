package yandex

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
)

const billingCloudBindingBindingResource = "yandex_billing_cloud_binding.test_cloud_binding_resource_binding"

func init() {
	resource.AddTestSweepers("yandex_billing_cloud_binding", &resource.Sweeper{
		Name: "yandex_billing_cloud_binding",
		F:    testSweepBillingCloudBinding,
		Dependencies: []string{
			"yandex_resourcemanager_cloud",
		},
	})
}

func testSweepBillingCloudBinding(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &billing.ListBillableObjectBindingsRequest{
		BillingAccountId: billingInstanceTestFirstBillingAccountId(),
	}
	it := conf.sdk.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(conf.Context(), req)
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

	return result.ErrorOrNil()
}

func sweepBillingCloudBinding(conf *Config, cloudId string) bool {
	return sweepWithRetry(sweepBillingCloudBindingOnce, conf, "Billing Cloud Binding", cloudId)
}

func sweepBillingCloudBindingOnce(conf *Config, instanceId string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexBillingServiceInstanceBindingDefaultTimeout)
	defer cancel()

	billableObject := &billing.BillableObject{
		Id:   instanceId,
		Type: billingCloudServiceInstanceBindingType,
	}
	req := &billing.BindBillableObjectRequest{
		BillingAccountId: billingInstanceTestSecondBillingAccountId(),
		BillableObject:   billableObject,
	}
	op, err := conf.sdk.Billing().BillingAccount().BindBillableObject(ctx, req)
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccResourceBillingCloudBinding_BindExistingCloudToExistingBillingAccount(t *testing.T) {
	firstBillingAccountId := billingInstanceTestFirstBillingAccountId()
	secondBillingAccountId := billingInstanceTestSecondBillingAccountId()
	cloudId := getExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBillingCloudBindingDestroy,
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
				ResourceName:      "yandex_billing_cloud_binding.test_cloud_binding_resource_binding",
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
	cloudId := getExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBillingCloudBindingDestroy,
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
				ResourceName:      "yandex_billing_cloud_binding.test_cloud_binding_resource_binding",
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
	cloudId := getExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBillingCloudBindingDestroy,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBillingCloudBindingDestroy,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckBillingCloudBindingDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceBillingCloudBindingBindCloudToBillingAccount(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Error while requesting API binding cloud to billing account"),
			},
		},
	})
}

func testAccCheckBillingCloudBindingDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_billing_cloud_binding" {
			continue
		}

		id, err := parseBindServiceInstanceId(rs.Primary.ID)

		if err != nil {
			return err
		}

		it := config.sdk.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
			context.Background(),
			&billing.ListBillableObjectBindingsRequest{
				BillingAccountId: billingInstanceTestFirstBillingAccountId(),
			},
		)

		for it.Next() {
			if it.Value().BillableObject.Type == billingCloudServiceInstanceBindingType &&
				it.Value().BillableObject.Id == id.serviceInstanceId {
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

		id, err := parseBindServiceInstanceId(rs.Primary.ID)

		if err != nil {
			return err
		}

		config := testAccProvider.Meta().(*Config)

		it := config.sdk.Billing().BillingAccount().BillingAccountBillableObjectBindingsIterator(
			context.Background(),
			&billing.ListBillableObjectBindingsRequest{
				BillingAccountId: id.billingAccountId,
			})

		for it.Next() {
			if it.Value().BillableObject.Type == billingCloudServiceInstanceBindingType &&
				it.Value().BillableObject.Id == id.serviceInstanceId {
				return nil
			}
		}

		return fmt.Errorf("Cloud bound to billing account not found")
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
