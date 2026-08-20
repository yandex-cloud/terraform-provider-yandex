package storage_bucket_iam_binding

import (
	"context"
	"testing"

	storage "github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	"google.golang.org/grpc"
)

type bucketGetterStub struct {
	request *storage.GetBucketRequest
	bucket  *storage.Bucket
}

func (s *bucketGetterStub) Get(_ context.Context, request *storage.GetBucketRequest, _ ...grpc.CallOption) (*storage.Bucket, error) {
	s.request = request
	return s.bucket, nil
}

func TestResolveBucketResourceIDUsesBucketName(t *testing.T) {
	client := &bucketGetterStub{
		bucket: &storage.Bucket{ResourceId: "resource-id"},
	}

	resourceID, err := resolveBucketResourceID(context.Background(), client, "bucket-name")
	if err != nil {
		t.Fatalf("resolve bucket resource ID: %v", err)
	}
	if resourceID != "resource-id" {
		t.Fatalf("expected resource ID %q, got %q", "resource-id", resourceID)
	}
	if client.request.GetName() != "bucket-name" {
		t.Fatalf("expected bucket name %q, got %q", "bucket-name", client.request.GetName())
	}
	if client.request.GetView() != storage.GetBucketRequest_VIEW_BASIC {
		t.Fatalf("expected basic bucket view, got %s", client.request.GetView())
	}
}
