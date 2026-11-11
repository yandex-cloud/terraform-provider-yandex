package pluginsdk_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_cdn "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cdn/pluginsdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetShieldingLocationWhenServiceIsUnavailable(t *testing.T) {
	t.Parallel()

	grpcServer := grpc.NewServer()
	endpointServer := &acceptanceenv.EndpointServer{}
	endpoint.RegisterApiEndpointServiceServer(grpcServer, endpointServer)

	listener := acceptanceenv.LoopbackListener(t)
	endpointServer.Endpoints = []*endpoint.ApiEndpoint{
		{Id: "endpoint", Address: listener.Addr().String()},
		{Id: "iam", Address: listener.Addr().String()},
		{Id: "cdn", Address: listener.Addr().String()},
	}
	go func() { _ = grpcServer.Serve(listener) }()
	defer grpcServer.Stop()

	config := providerconfig.Config{
		Endpoint:  listener.Addr().String(),
		Token:     "t1.a.b",
		Plaintext: true,
	}
	require.NoError(t, config.InitAndValidate(context.Background(), acceptanceenv.TerraformVersion, false))

	locationID, err := svc_cdn.GetShieldingLocation(context.Background(), "resource-id", &config)
	require.NoError(t, err)
	assert.Nil(t, locationID)
}

func TestIsIgnorableShieldingReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "shielding is not configured",
			err:  status.Error(codes.NotFound, "shielding is not configured"),
			want: true,
		},
		{
			name: "shielding service is not available",
			err:  status.Error(codes.Unimplemented, "unknown service yandex.cloud.cdn.v1.ShieldingService"),
			want: true,
		},
		{
			name: "wrapped unimplemented response",
			err: fmt.Errorf(
				"reading shielding: %w",
				status.Error(codes.Unimplemented, "unknown shielding service"),
			),
			want: true,
		},
		{
			name: "permission denied",
			err:  status.Error(codes.PermissionDenied, "permission denied"),
			want: false,
		},
		{
			name: "service unavailable",
			err:  status.Error(codes.Unavailable, "service unavailable"),
			want: false,
		},
		{
			name: "non grpc error",
			err:  errors.New("network failure"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := svc_cdn.IsIgnorableShieldingReadError(tt.err); got != tt.want {
				t.Fatalf("isIgnorableShieldingReadError() = %v, want %v", got, tt.want)
			}
		})
	}
}
