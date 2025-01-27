package billing_cloud_binding_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const billingCloudBindingBindingDataSource = "data.yandex_billing_cloud_binding.test_cloud_binding_data_binding"

func TestAccDataSourceBillingCloudBinding_BindExistingCloudToExistingBillingAccountThenCheckData(t *testing.T) {
	firstBillingAccountId := billingInstanceTestFirstBillingAccountId()
	secondBillingAccountId := billingInstanceTestSecondBillingAccountId()
	cloudId := test.GetExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			return testAccCheckBillingCloudBindingDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(firstBillingAccountId, cloudId),
			},
			{
				Config: testAccDataSourceBillingCloudBindingGetDataSource(firstBillingAccountId, cloudId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(billingCloudBindingBindingDataSource, "billing_account_id", firstBillingAccountId),
					resource.TestCheckResourceAttr(billingCloudBindingBindingDataSource, "cloud_id", cloudId),
				),
			},
			{
				Config: testAccResourceBillingCloudBindingBindCloudToBillingAccount(secondBillingAccountId, cloudId),
			},
		},
	})
}

func TestAccDataSourceBillingCloudBinding_CheckNonExistingBillingAccountData(t *testing.T) {
	billingAccountId := fmt.Sprintf("non-existing-billing-account-id-%s", acctest.RandString(10))
	cloudId := test.GetExampleCloudID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceBillingCloudBindingGetDataSource(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Bound cloud to billing account not found"),
			},
		},
	})
}

func TestAccDataSourceBillingCloudBinding_CheckNonExistingCloudData(t *testing.T) {
	billingAccountId := billingInstanceTestFirstBillingAccountId()
	cloudId := fmt.Sprintf("non-existing-cloud-id-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceBillingCloudBindingGetDataSource(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Bound cloud to billing account not found"),
			},
		},
	})
}

func TestAccDataSourceBillingCloudBinding_CheckNonExistingBillingAccountNonExistingCloudData(t *testing.T) {
	billingAccountId := fmt.Sprintf("non-existing-billing-account-id-%s", acctest.RandString(10))
	cloudId := fmt.Sprintf("non-existing-cloud-id-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceBillingCloudBindingGetDataSource(billingAccountId, cloudId),
				ExpectError: regexp.MustCompile("Bound cloud to billing account not found"),
			},
		},
	})
}

func testAccDataSourceBillingCloudBindingGetDataSource(billingAccountId string, cloudId string) string {
	return fmt.Sprintf(`
data "yandex_billing_cloud_binding" "test_cloud_binding_data_binding" {
	billing_account_id = "%s"
	cloud_id = "%s"
}`, billingAccountId, cloudId)
}
