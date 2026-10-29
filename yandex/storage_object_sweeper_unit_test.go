package yandex

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
)

type storageSweepFake struct {
	list func(*awsS3.ListObjectVersionsInput) (*awsS3.ListObjectVersionsOutput, error)
	head func(*awsS3.HeadObjectInput) (*awsS3.HeadObjectOutput, error)
	hold func(*awsS3.PutObjectLegalHoldInput) error
	del  func(*awsS3.DeleteObjectInput) error
}

func (f storageSweepFake) ListObjectVersionsWithContext(_ aws.Context, input *awsS3.ListObjectVersionsInput, _ ...request.Option) (*awsS3.ListObjectVersionsOutput, error) {
	return f.list(input)
}

func (f storageSweepFake) HeadObjectWithContext(_ aws.Context, input *awsS3.HeadObjectInput, _ ...request.Option) (*awsS3.HeadObjectOutput, error) {
	return f.head(input)
}

func (f storageSweepFake) PutObjectLegalHoldWithContext(_ aws.Context, input *awsS3.PutObjectLegalHoldInput, _ ...request.Option) (*awsS3.PutObjectLegalHoldOutput, error) {
	return &awsS3.PutObjectLegalHoldOutput{}, f.hold(input)
}

func (f storageSweepFake) DeleteObjectWithContext(_ aws.Context, input *awsS3.DeleteObjectInput, _ ...request.Option) (*awsS3.DeleteObjectOutput, error) {
	return &awsS3.DeleteObjectOutput{}, f.del(input)
}

// Ensure the real SDK client satisfies the same contract as the fake.
var _ storageObjectSweepClient = (*awsS3.S3)(nil)

func TestStorageObjectSweepLocks(t *testing.T) {
	denied := errors.New("AccessDenied")
	future := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name       string
		mode       string
		legalHold  string
		until      *time.Time
		fail       string
		wantCalls  []string
		wantBypass bool
		wantError  string
	}{
		{name: "unlocked", wantCalls: []string{"head", "delete"}},
		{name: "governance", mode: awsS3.ObjectLockModeGovernance, until: &future, wantCalls: []string{"head", "delete"}, wantBypass: true},
		{name: "legal hold", legalHold: awsS3.ObjectLockLegalHoldStatusOn, wantCalls: []string{"head", "hold", "delete"}},
		{name: "both locks", mode: awsS3.ObjectLockModeGovernance, legalHold: awsS3.ObjectLockLegalHoldStatusOn, until: &future, wantCalls: []string{"head", "hold", "delete"}, wantBypass: true},
		{name: "active compliance", mode: awsS3.ObjectLockModeCompliance, legalHold: awsS3.ObjectLockLegalHoldStatusOn, until: &future, wantCalls: []string{"head"}, wantError: "COMPLIANCE retention prevents deletion until 2100-01-01T00:00:00Z"},
		{name: "expired compliance", mode: awsS3.ObjectLockModeCompliance, until: &past, wantCalls: []string{"head", "delete"}},
		{name: "head denied", fail: "head", wantCalls: []string{"head"}, wantError: "reading lock state"},
		{name: "hold denied", legalHold: awsS3.ObjectLockLegalHoldStatusOn, fail: "hold", wantCalls: []string{"head", "hold"}, wantError: "removing legal hold"},
		{name: "delete denied", fail: "delete", wantCalls: []string{"head", "delete"}, wantError: "deleting version"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			checkVersion := func(bucket, key, version *string) {
				t.Helper()
				require.Equal(t, "tf-test-locks", aws.StringValue(bucket))
				require.Equal(t, "key", aws.StringValue(key))
				require.Equal(t, "old-version", aws.StringValue(version))
			}
			client := storageSweepFake{
				head: func(in *awsS3.HeadObjectInput) (*awsS3.HeadObjectOutput, error) {
					calls = append(calls, "head")
					checkVersion(in.Bucket, in.Key, in.VersionId)
					if tt.fail == "head" {
						return nil, denied
					}
					return &awsS3.HeadObjectOutput{ObjectLockMode: aws.String(tt.mode), ObjectLockLegalHoldStatus: aws.String(tt.legalHold), ObjectLockRetainUntilDate: tt.until}, nil
				},
				hold: func(in *awsS3.PutObjectLegalHoldInput) error {
					calls = append(calls, "hold")
					checkVersion(in.Bucket, in.Key, in.VersionId)
					require.Equal(t, awsS3.ObjectLockLegalHoldStatusOff, aws.StringValue(in.LegalHold.Status))
					if tt.fail == "hold" {
						return denied
					}
					return nil
				},
				del: func(in *awsS3.DeleteObjectInput) error {
					calls = append(calls, "delete")
					checkVersion(in.Bucket, in.Key, in.VersionId)
					require.Equal(t, tt.wantBypass, aws.BoolValue(in.BypassGovernanceRetention))
					if !tt.wantBypass {
						require.Nil(t, in.BypassGovernanceRetention, "unnecessary bypass header")
					}
					if tt.fail == "delete" {
						return denied
					}
					return nil
				},
			}
			err := sweepStorageObjectVersion(context.Background(), "tf-test-locks", &awsS3.ObjectVersion{Key: aws.String("key"), VersionId: aws.String("old-version")}, client)
			if tt.wantError == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantError)
			}
			if tt.fail != "" {
				require.ErrorIs(t, err, denied)
			}
			require.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestStorageObjectSweepPaginationAndScope(t *testing.T) {
	var listed, deleted, headed []string
	client := storageSweepFake{
		list: func(in *awsS3.ListObjectVersionsInput) (*awsS3.ListObjectVersionsOutput, error) {
			listed = append(listed, aws.StringValue(in.Bucket))
			if len(listed) == 1 {
				require.Nil(t, in.KeyMarker, "first page has a key marker")
				require.Nil(t, in.VersionIdMarker, "first page has a version marker")
				return &awsS3.ListObjectVersionsOutput{
					Versions:    []*awsS3.ObjectVersion{{Key: aws.String("same-key"), VersionId: aws.String("v1")}},
					IsTruncated: aws.Bool(true), NextKeyMarker: aws.String("same-key"), NextVersionIdMarker: aws.String("v1"),
				}, nil
			}
			require.Len(t, listed, 2)
			require.Equal(t, "same-key", aws.StringValue(in.KeyMarker))
			require.Equal(t, "v1", aws.StringValue(in.VersionIdMarker))
			return &awsS3.ListObjectVersionsOutput{
				Versions:      []*awsS3.ObjectVersion{{Key: aws.String("same-key"), VersionId: aws.String("v2")}},
				DeleteMarkers: []*awsS3.DeleteMarkerEntry{{Key: aws.String("same-key"), VersionId: aws.String("marker")}},
			}, nil
		},
		head: func(in *awsS3.HeadObjectInput) (*awsS3.HeadObjectOutput, error) {
			headed = append(headed, aws.StringValue(in.VersionId))
			return &awsS3.HeadObjectOutput{}, nil
		},
		del: func(in *awsS3.DeleteObjectInput) error {
			require.Equal(t, "tf-test-pagination", aws.StringValue(in.Bucket))
			require.Equal(t, "same-key", aws.StringValue(in.Key))
			require.Nil(t, in.BypassGovernanceRetention)
			deleted = append(deleted, aws.StringValue(in.VersionId))
			return nil
		},
	}
	err := sweepStorageObjects(context.Background(), []*awsS3.Bucket{
		{Name: aws.String("production")}, {Name: aws.String("tf-test-pagination")}, {Name: aws.String("other-tf-test-bucket")},
	}, client)
	require.NoError(t, err)
	require.Equal(t, []string{"tf-test-pagination", "tf-test-pagination"}, listed)
	require.Equal(t, []string{"v1", "v2"}, headed)
	require.Equal(t, []string{"v1", "v2", "marker"}, deleted)
}

func TestStorageObjectSweepErrorsContinue(t *testing.T) {
	denied := errors.New("AccessDenied")
	var listed, deleted []string
	client := storageSweepFake{
		list: func(in *awsS3.ListObjectVersionsInput) (*awsS3.ListObjectVersionsOutput, error) {
			listed = append(listed, aws.StringValue(in.Bucket))
			if aws.StringValue(in.Bucket) == "tf-test-list-error" {
				return nil, denied
			}
			return &awsS3.ListObjectVersionsOutput{
				Versions:      []*awsS3.ObjectVersion{{Key: aws.String("key"), VersionId: aws.String("failed")}, {Key: aws.String("key"), VersionId: aws.String("ok")}},
				DeleteMarkers: []*awsS3.DeleteMarkerEntry{{Key: aws.String("key"), VersionId: aws.String("marker")}},
			}, nil
		},
		head: func(_ *awsS3.HeadObjectInput) (*awsS3.HeadObjectOutput, error) { return &awsS3.HeadObjectOutput{}, nil },
		del: func(in *awsS3.DeleteObjectInput) error {
			deleted = append(deleted, aws.StringValue(in.VersionId))
			if aws.StringValue(in.VersionId) != "ok" {
				return denied
			}
			return nil
		},
	}
	err := sweepStorageObjects(context.Background(), []*awsS3.Bucket{{Name: aws.String("tf-test-list-error")}, {Name: aws.String("tf-test-next")}}, client)
	require.ErrorIs(t, err, denied)
	for _, want := range []string{"listing versions", "tf-test-list-error", `key "key" version "failed"`, `deleting marker`, `version "marker"`} {
		require.ErrorContains(t, err, want)
	}
	require.Len(t, listed, 2)
	require.Equal(t, []string{"failed", "ok", "marker"}, deleted)
}

func TestStorageObjectSweepPaginationDoesNotAdvance(t *testing.T) {
	calls := 0
	client := storageSweepFake{list: func(_ *awsS3.ListObjectVersionsInput) (*awsS3.ListObjectVersionsOutput, error) {
		calls++
		return &awsS3.ListObjectVersionsOutput{IsTruncated: aws.Bool(true), NextKeyMarker: aws.String("key"), NextVersionIdMarker: aws.String("v1")}, nil
	}}
	err := sweepStorageObjectVersions(context.Background(), "tf-test-stuck", client)
	require.ErrorContains(t, err, "pagination did not advance")
	require.Equal(t, 2, calls)
}
