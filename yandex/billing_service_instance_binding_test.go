package yandex

import (
	"os"
	"time"
)

func billingInstanceTestFirstBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_1")
}
func billingInstanceTestSecondBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_2")
}

const yandexBillingServiceInstanceBindingDefaultTimeout = 1 * time.Minute
