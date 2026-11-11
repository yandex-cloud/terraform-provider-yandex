package storage_bucket_policy_test

import (
	storagesdk "github.com/yandex-cloud/go-sdk/services/storage/v1"
	framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	fixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/testfixtures"
)

func bucketClient() fixtures.BucketClient {
	return storagesdk.NewBucketClient(test.AccProvider.(*framework.Provider).GetConfig().SDKv2)
}
