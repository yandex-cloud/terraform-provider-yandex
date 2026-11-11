package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BucketExists(client func() BucketClient, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := client().Get(context.Background(), &storage.GetBucketRequest{
			Name: name,
			View: storage.GetBucketRequest_VIEW_BASIC,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func AccCheckBucketDestroy(client func() BucketClient, name string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		_, err := client().Get(context.Background(), &storage.GetBucketRequest{
			Name: name,
			View: storage.GetBucketRequest_VIEW_BASIC,
		})
		if err == nil {
			return fmt.Errorf("bucket still exists")
		}

		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("check bucket %q deletion: %w", name, err)
		}
		return nil
	}
}

type BucketClient interface {
	Get(context.Context, *storage.GetBucketRequest, ...grpc.CallOption) (*storage.Bucket, error)
}
