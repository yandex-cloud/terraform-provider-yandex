package yandex

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetShieldingLocationWhenServiceIsUnavailable(t *testing.T) {
	t.Parallel()

	grpcServer := grpc.NewServer()
	endpointServer := &userAgentMockServerAPIEndpoint{}
	endpoint.RegisterApiEndpointServiceServer(grpcServer, endpointServer)

	listener := localListener(t)
	endpointServer.endpoints = []*endpoint.ApiEndpoint{
		{Id: "endpoint", Address: listener.Addr().String()},
		{Id: "iam", Address: listener.Addr().String()},
		{Id: "cdn", Address: listener.Addr().String()},
	}
	go func() { _ = grpcServer.Serve(listener) }()
	defer grpcServer.Stop()

	config := Config{
		Endpoint:  listener.Addr().String(),
		Token:     "t1.a.b",
		Plaintext: true,
	}
	require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))

	locationID, err := getShieldingLocation(context.Background(), "resource-id", &config)
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

			if got := isIgnorableShieldingReadError(tt.err); got != tt.want {
				t.Fatalf("isIgnorableShieldingReadError() = %v, want %v", got, tt.want)
			}
		})
	}
}
