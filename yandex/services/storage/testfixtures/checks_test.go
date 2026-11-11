package testfixtures

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type bucketClientStub struct {
	name string
	view storage.GetBucketRequest_View
	err  error
}

func (c *bucketClientStub) Get(_ context.Context, req *storage.GetBucketRequest, _ ...grpc.CallOption) (*storage.Bucket, error) {
	c.name, c.view = req.Name, req.View
	return &storage.Bucket{}, c.err
}

func TestBucketDestroyErrors(t *testing.T) {
	for _, tc := range []struct {
		name      string
		apiError  error
		wantError bool
	}{
		{"exists", nil, true},
		{"deleted", status.Error(codes.NotFound, "missing"), false},
		{"permission denied", status.Error(codes.PermissionDenied, "denied"), true},
		{"unavailable", status.Error(codes.Unavailable, "unavailable"), true},
		{"other", errors.New("transport failure"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &bucketClientStub{err: tc.apiError}
			calls := 0
			check := AccCheckBucketDestroy(func() BucketClient { calls++; return client }, "test-bucket")
			if calls != 0 {
				t.Fatal("client must be resolved lazily")
			}
			err := check(nil)
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError %v", err, tc.wantError)
			}
			if client.name != "test-bucket" || client.view != storage.GetBucketRequest_VIEW_BASIC {
				t.Fatalf("unexpected request: %+v", client)
			}
			if tc.apiError != nil && tc.wantError && (!errors.Is(err, tc.apiError) || !strings.Contains(err.Error(), "test-bucket")) {
				t.Fatalf("error lost cause or bucket: %v", err)
			}
		})
	}
}
