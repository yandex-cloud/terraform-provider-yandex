package testfixtures

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func destroyState(resourceType string) *terraform.State {
	return &terraform.State{Modules: []*terraform.ModuleState{{
		Path: []string{"root"},
		Resources: map[string]*terraform.ResourceState{
			"tested": {Type: resourceType, Primary: &terraform.InstanceState{ID: "resource-id"}},
		},
	}}}
}

type imageGetFunc func(context.Context, *compute.GetImageRequest, ...grpc.CallOption) (*compute.Image, error)

func (get imageGetFunc) Get(ctx context.Context, req *compute.GetImageRequest, opts ...grpc.CallOption) (*compute.Image, error) {
	return get(ctx, req, opts...)
}

func TestCheckImageDestroy(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		response     *compute.Image
		apiError     error
		wantError    bool
		wantCall     bool
	}{
		{name: "deleted", resourceType: "yandex_compute_image", apiError: status.Error(codes.NotFound, "not found"), wantCall: true},
		{name: "exists", resourceType: "yandex_compute_image", response: &compute.Image{}, wantError: true, wantCall: true},
		{name: "permission denied", resourceType: "yandex_compute_image", apiError: status.Error(codes.PermissionDenied, "denied"), wantError: true, wantCall: true},
		{name: "transport error", resourceType: "yandex_compute_image", apiError: status.Error(codes.Unavailable, "connection unavailable"), wantError: true, wantCall: true},
		{name: "unrelated resource", resourceType: "unrelated_resource"},
		{name: "standard image", resourceType: "yandex_compute_image", response: &compute.Image{FolderId: "standard-images"}, wantCall: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			get := imageGetFunc(func(_ context.Context, req *compute.GetImageRequest, _ ...grpc.CallOption) (*compute.Image, error) {
				called = true
				if req.ImageId != "resource-id" {
					t.Fatalf("requested id = %q; want resource-id", req.ImageId)
				}
				return tc.response, tc.apiError
			})
			err := checkImageDestroy(get, destroyState(tc.resourceType))
			if (err != nil) != tc.wantError {
				t.Fatalf("destroy check error = %v; wantError %v", err, tc.wantError)
			}
			if tc.wantError && tc.apiError != nil && !errors.Is(err, tc.apiError) {
				t.Fatalf("destroy check did not preserve API error: %v", err)
			}
			if called != tc.wantCall {
				t.Fatalf("Get called = %v; want %v", called, tc.wantCall)
			}
		})
	}
}
