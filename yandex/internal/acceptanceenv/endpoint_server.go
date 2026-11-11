package acceptanceenv

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
)

type EndpointServer struct {
	UserAgent string
	Endpoints []*endpoint.ApiEndpoint
	ListError error
	ListCalls int
}

func (s *EndpointServer) Get(
	context.Context,
	*endpoint.GetApiEndpointRequest,
) (*endpoint.ApiEndpoint, error) {
	return &endpoint.ApiEndpoint{}, nil
}

func (s *EndpointServer) List(
	ctx context.Context,
	r *endpoint.ListApiEndpointsRequest,
) (*endpoint.ListApiEndpointsResponse, error) {
	s.ListCalls++
	reqMd, _ := metadata.FromIncomingContext(ctx)
	userAgent := reqMd.Get("user-agent")
	if len(userAgent) > 0 {
		s.UserAgent = userAgent[0]
	}

	if s.ListError != nil {
		return nil, s.ListError
	}

	return &endpoint.ListApiEndpointsResponse{Endpoints: s.Endpoints}, nil
}

func LoopbackListener(t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		l, err = net.Listen("tcp6", "[::1]:0")
	}
	require.NoError(t, err, "failed to listen on any port")
	return l
}
