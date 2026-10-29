package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-multierror"
)

type storageObjectSweepClient interface {
	ListObjectVersionsWithContext(aws.Context, *awsS3.ListObjectVersionsInput, ...request.Option) (*awsS3.ListObjectVersionsOutput, error)
	HeadObjectWithContext(aws.Context, *awsS3.HeadObjectInput, ...request.Option) (*awsS3.HeadObjectOutput, error)
	PutObjectLegalHoldWithContext(aws.Context, *awsS3.PutObjectLegalHoldInput, ...request.Option) (*awsS3.PutObjectLegalHoldOutput, error)
	DeleteObjectWithContext(aws.Context, *awsS3.DeleteObjectInput, ...request.Option) (*awsS3.DeleteObjectOutput, error)
}

// Lock removal is limited to test cleanup; it must not change resource destroy semantics.
func sweepStorageObjects(ctx context.Context, buckets []*awsS3.Bucket, client storageObjectSweepClient) error {
	var result *multierror.Error
	for _, bucket := range buckets {
		name := aws.StringValue(bucket.Name)
		if !strings.HasPrefix(name, "tf-test-") {
			continue
		}
		if err := sweepStorageObjectVersions(ctx, name, client); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

func sweepStorageObjectVersions(ctx context.Context, bucket string, client storageObjectSweepClient) error {
	var result *multierror.Error
	input := &awsS3.ListObjectVersionsInput{Bucket: aws.String(bucket)}
	for {
		page, err := client.ListObjectVersionsWithContext(ctx, input)
		if err != nil {
			return multierror.Append(result, fmt.Errorf("listing versions in bucket %q: %w", bucket, err)).ErrorOrNil()
		}
		for _, version := range page.Versions {
			if err := sweepStorageObjectVersion(ctx, bucket, version, client); err != nil {
				result = multierror.Append(result, fmt.Errorf("cleaning bucket %q key %q version %q: %w",
					bucket, aws.StringValue(version.Key), aws.StringValue(version.VersionId), err))
			}
		}
		for _, marker := range page.DeleteMarkers {
			_, err := client.DeleteObjectWithContext(ctx, &awsS3.DeleteObjectInput{
				Bucket: input.Bucket, Key: marker.Key, VersionId: marker.VersionId,
			})
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("deleting marker in bucket %q key %q version %q: %w",
					bucket, aws.StringValue(marker.Key), aws.StringValue(marker.VersionId), err))
			}
		}
		if !aws.BoolValue(page.IsTruncated) {
			return result.ErrorOrNil()
		}
		if aws.StringValue(page.NextKeyMarker) == "" ||
			(aws.StringValue(page.NextKeyMarker) == aws.StringValue(input.KeyMarker) &&
				aws.StringValue(page.NextVersionIdMarker) == aws.StringValue(input.VersionIdMarker)) {
			return multierror.Append(result, fmt.Errorf("listing versions in bucket %q: pagination did not advance", bucket)).ErrorOrNil()
		}
		input.KeyMarker = page.NextKeyMarker
		input.VersionIdMarker = page.NextVersionIdMarker
	}
}

func sweepStorageObjectVersion(ctx context.Context, bucket string, version *awsS3.ObjectVersion, client storageObjectSweepClient) error {
	head, err := client.HeadObjectWithContext(ctx, &awsS3.HeadObjectInput{
		Bucket: aws.String(bucket), Key: version.Key, VersionId: version.VersionId,
	})
	if err != nil {
		return fmt.Errorf("reading lock state: %w", err)
	}
	if aws.StringValue(head.ObjectLockMode) == awsS3.ObjectLockModeCompliance &&
		aws.TimeValue(head.ObjectLockRetainUntilDate).After(time.Now()) {
		return fmt.Errorf("COMPLIANCE retention prevents deletion until %s", head.ObjectLockRetainUntilDate.Format(time.RFC3339))
	}
	if aws.StringValue(head.ObjectLockLegalHoldStatus) == awsS3.ObjectLockLegalHoldStatusOn {
		_, err := client.PutObjectLegalHoldWithContext(ctx, &awsS3.PutObjectLegalHoldInput{
			Bucket: aws.String(bucket), Key: version.Key, VersionId: version.VersionId,
			LegalHold: &awsS3.ObjectLockLegalHold{Status: aws.String(awsS3.ObjectLockLegalHoldStatusOff)},
		})
		if err != nil {
			return fmt.Errorf("removing legal hold: %w", err)
		}
	}
	input := &awsS3.DeleteObjectInput{Bucket: aws.String(bucket), Key: version.Key, VersionId: version.VersionId}
	if aws.StringValue(head.ObjectLockMode) == awsS3.ObjectLockModeGovernance {
		input.BypassGovernanceRetention = aws.Bool(true)
	}
	if _, err := client.DeleteObjectWithContext(ctx, input); err != nil {
		return fmt.Errorf("deleting version: %w", err)
	}
	return nil
}
