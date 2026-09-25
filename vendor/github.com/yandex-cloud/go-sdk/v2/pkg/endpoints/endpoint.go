package endpoints

import (
	"google.golang.org/grpc"
)

// Endpoint represents a gRPC connection target with specific connection options.
type Endpoint struct {
	// Addr is endpoint address in host:port format.
	Addr string
	// DialOptions is Endpoint specific connect options.
	// Typical options:
	// * 'plaintext': grpc.WithTransportCredentials(insecure.NewCredentials())
	// * 'insecure': grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ InsecureSkipVerify: true }))
	DialOptions []grpc.DialOption
}
