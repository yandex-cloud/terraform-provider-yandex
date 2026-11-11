package sweep

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	svc_object_storage "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/pluginsdk"
)

func registerResourceYandexStorageBucket(add func(string, *resource.Sweeper)) {
	add("yandex_storage_bucket", &resource.Sweeper{
		Name: "yandex_storage_bucket",
		F:    sweepAllStorageBucket,
		Dependencies: []string{
			"yandex_storage_object",
		},
	})
}

func sweepAllStorageBucket(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}
	s3Client, err := svc_object_storage.GetS3ClientByKeys(context.TODO(), "", "", conf)
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("error creating storage client: %s", err))
		return result.ErrorOrNil()
	}

	buckets, err := s3Client.S3().ListBuckets(&awsS3.ListBucketsInput{})
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to list storage buckets: %s", err))
		return result.ErrorOrNil()
	}

	return SweepStorageBuckets(context.TODO(), buckets.Buckets, s3Client.DeleteBucket)
}

func SweepStorageBuckets(
	ctx context.Context,
	buckets []*awsS3.Bucket,
	deleteBucket func(context.Context, string, bool) error,
) error {
	result := &multierror.Error{}
	for _, b := range buckets {
		bucketName := aws.StringValue(b.Name)
		if !strings.HasPrefix(bucketName, "tf-test-") {
			continue
		}

		if err := deleteBucket(ctx, bucketName, true); err != nil {
			result = multierror.Append(
				result,
				fmt.Errorf("failed to delete bucket: %s, error: %s", bucketName, err),
			)
		}
	}

	return result.ErrorOrNil()
}
