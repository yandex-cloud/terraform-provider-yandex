package sweep

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/billing/v1"
	billingsdk "github.com/yandex-cloud/go-sdk/services/billing/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const billingCloudServiceInstanceBindingType = "cloud"

func billingInstanceTestFirstBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_1")
}

func billingInstanceTestSecondBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_2")
}

const yandexBillingServiceInstanceBindingDefaultTimeout = 1 * time.Minute

func registerBillingCloudBindingResource(add func(string, *resource.Sweeper)) {
	add("yandex_billing_cloud_binding", &resource.Sweeper{
		Name: "yandex_billing_cloud_binding",
		F:    sweepAllBillingCloudBinding,
		Dependencies: []string{
			"yandex_resourcemanager_cloud",
		},
	})
}

func sweepAllBillingCloudBinding(_ string) error {
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
	it := billingsdk.NewBillingAccountClient(conf.SDKv2).BillableObjectBindingsIterator(context.Background(), req)
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
	op, err := billingsdk.NewBillingAccountClient(conf.SDKv2).BindBillableObject(ctx, req)
	return test.HandleSweepOperation(ctx, op, err)
}
