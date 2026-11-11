package testfixtures

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
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

type networkGetFunc func(context.Context, *vpc.GetNetworkRequest, ...grpc.CallOption) (*vpc.Network, error)

func (get networkGetFunc) Get(ctx context.Context, req *vpc.GetNetworkRequest, opts ...grpc.CallOption) (*vpc.Network, error) {
	return get(ctx, req, opts...)
}

func TestCheckNetworkDestroy(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		response     *vpc.Network
		apiError     error
		wantError    bool
		wantCall     bool
	}{
		{name: "deleted", resourceType: "yandex_vpc_network", apiError: status.Error(codes.NotFound, "not found"), wantCall: true},
		{name: "exists", resourceType: "yandex_vpc_network", response: &vpc.Network{}, wantError: true, wantCall: true},
		{name: "permission denied", resourceType: "yandex_vpc_network", apiError: status.Error(codes.PermissionDenied, "denied"), wantError: true, wantCall: true},
		{name: "transport error", resourceType: "yandex_vpc_network", apiError: status.Error(codes.Unavailable, "connection unavailable"), wantError: true, wantCall: true},
		{name: "unrelated resource", resourceType: "unrelated_resource"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			get := networkGetFunc(func(_ context.Context, req *vpc.GetNetworkRequest, _ ...grpc.CallOption) (*vpc.Network, error) {
				called = true
				if req.NetworkId != "resource-id" {
					t.Fatalf("requested id = %q; want resource-id", req.NetworkId)
				}
				return tc.response, tc.apiError
			})
			err := checkNetworkDestroy(get, destroyState(tc.resourceType))
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

type addressGetFunc func(context.Context, *vpc.GetAddressRequest, ...grpc.CallOption) (*vpc.Address, error)

func (get addressGetFunc) Get(ctx context.Context, req *vpc.GetAddressRequest, opts ...grpc.CallOption) (*vpc.Address, error) {
	return get(ctx, req, opts...)
}

func TestCheckAddressDestroy(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		response     *vpc.Address
		apiError     error
		wantError    bool
		wantCall     bool
	}{
		{name: "deleted", resourceType: "yandex_vpc_address", apiError: status.Error(codes.NotFound, "not found"), wantCall: true},
		{name: "exists", resourceType: "yandex_vpc_address", response: &vpc.Address{}, wantError: true, wantCall: true},
		{name: "permission denied", resourceType: "yandex_vpc_address", apiError: status.Error(codes.PermissionDenied, "denied"), wantError: true, wantCall: true},
		{name: "transport error", resourceType: "yandex_vpc_address", apiError: status.Error(codes.Unavailable, "connection unavailable"), wantError: true, wantCall: true},
		{name: "unrelated resource", resourceType: "unrelated_resource"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			get := addressGetFunc(func(_ context.Context, req *vpc.GetAddressRequest, _ ...grpc.CallOption) (*vpc.Address, error) {
				called = true
				if req.AddressId != "resource-id" {
					t.Fatalf("requested id = %q; want resource-id", req.AddressId)
				}
				return tc.response, tc.apiError
			})
			err := checkAddressDestroy(get, destroyState(tc.resourceType))
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
