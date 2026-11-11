package testfixtures

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type symmetricKeyGetterFunc func(context.Context, *kms.GetSymmetricKeyRequest) (*kms.SymmetricKey, error)

func (f symmetricKeyGetterFunc) Get(ctx context.Context, req *kms.GetSymmetricKeyRequest, _ ...grpc.CallOption) (*kms.SymmetricKey, error) {
	return f(ctx, req)
}

func TestCheckSymmetricKeysDestroy(t *testing.T) {
	transportErr := errors.New("transport failed")
	for _, tt := range []struct {
		name      string
		getErr    error
		wantError bool
	}{
		{name: "key exists", wantError: true},
		{name: "key deleted", getErr: status.Error(codes.NotFound, "deleted")},
		{name: "permission denied", getErr: status.Error(codes.PermissionDenied, "forbidden"), wantError: true},
		{name: "transport failure", getErr: transportErr, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckSymmetricKeysDestroy(func() (SymmetricKeyGetter, error) {
				return symmetricKeyGetterFunc(func(_ context.Context, req *kms.GetSymmetricKeyRequest) (*kms.SymmetricKey, error) {
					if req.KeyId != "key-id" {
						t.Fatalf("unexpected key ID: %q", req.KeyId)
					}
					return &kms.SymmetricKey{}, tt.getErr
				}), nil
			})
			state := &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{"key": {Type: "yandex_kms_symmetric_key", Primary: &terraform.InstanceState{ID: "key-id"}}}}}}
			err := check(state)
			if (err != nil) != tt.wantError {
				t.Fatalf("check returned %v; want error: %v", err, tt.wantError)
			}
			if err != nil && !strings.Contains(err.Error(), "key-id") {
				t.Fatalf("error omits key ID: %v", err)
			}
			if tt.getErr != nil && tt.wantError && !errors.Is(err, tt.getErr) {
				t.Fatalf("error does not preserve API error: %v", err)
			}
		})
	}
}

func TestCheckSymmetricKeysDestroyClientError(t *testing.T) {
	expected := errors.New("provider is not configured")
	check := CheckSymmetricKeysDestroy(func() (SymmetricKeyGetter, error) { return nil, expected })
	state := &terraform.State{Modules: []*terraform.ModuleState{{Path: []string{"root"}, Resources: map[string]*terraform.ResourceState{"key": {Type: "yandex_kms_symmetric_key", Primary: &terraform.InstanceState{ID: "key-id"}}}}}}
	if err := check(state); !errors.Is(err, expected) {
		t.Fatalf("check returned %v; want client error", err)
	}
}
