package testhelpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func BucketExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := AccProvider.(*yandex_framework.Provider).GetConfig()

		_, err := config.SDK.StorageAPI().Bucket().Get(context.Background(), &storage.GetBucketRequest{
			Name: name,
			View: storage.GetBucketRequest_VIEW_BASIC,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func AccCheckBucketDestroy(name string) func(*terraform.State) error {
	return func(s *terraform.State) error {

		config := AccProvider.(*yandex_framework.Provider).GetConfig()

		_, err := config.SDK.StorageAPI().Bucket().Get(context.Background(), &storage.GetBucketRequest{
			Name: name,
			View: storage.GetBucketRequest_VIEW_BASIC,
		})
		if err == nil {
			return fmt.Errorf("bucket still exists")
		}

		return nil
	}
}
