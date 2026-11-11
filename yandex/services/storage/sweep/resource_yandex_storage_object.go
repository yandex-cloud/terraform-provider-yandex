package sweep

import (
	"context"
	"fmt"

	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	svc_object_storage "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/pluginsdk"
)

func registerResourceYandexStorageObject(add func(string, *resource.Sweeper)) {
	add("yandex_storage_object", &resource.Sweeper{
		Name:         "yandex_storage_object",
		F:            SweepObjects,
		Dependencies: []string{},
	})
}

func SweepObjects(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	ctx := context.Background()
	s3Client, err := svc_object_storage.GetS3ClientByKeys(ctx, "", "", conf)
	if err != nil {
		return fmt.Errorf("error creating storage client: %w", err)
	}

	buckets, err := s3Client.S3().ListBucketsWithContext(ctx, &awsS3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("failed to list storage buckets: %w", err)
	}

	return SweepStorageObjects(ctx, buckets.Buckets, s3Client.S3())
}
