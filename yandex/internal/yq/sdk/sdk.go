package sdk

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// SDK is Yandex Query SDK
type SDK struct {
	conf   Config
	client client.YQClient
}

func validateConfig(conf *Config) bool {
	return conf.FolderID != "" && conf.Endpoint != ""
}

func (sdk *SDK) Client() client.YQClient {
	return sdk.client
}

func NewYQSDK(ctx context.Context, conf Config, opts ...grpc.DialOption) (*SDK, error) {
	if !validateConfig(&conf) {
		// no error here, YQ is optional
		return nil, nil
	}

	mdMiddleware := newYQMDMiddleware(conf.AuthTokenProvider, conf.FolderID)

	dialOpts := make([]grpc.DialOption, 0, len(opts)+3) // 2 for interceptors, 1 for tls
	dialOpts = append(dialOpts,
		grpc.WithChainUnaryInterceptor(mdMiddleware.unaryInterceptor),
		grpc.WithChainStreamInterceptor(mdMiddleware.streamInterceptor),
	)

	if conf.Plaintext {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := conf.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}
		creds := credentials.NewTLS(tlsConfig)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	}

	dialOpts = append(dialOpts, opts...)

	grpcConn, err := grpc.NewClient(conf.Endpoint, dialOpts...)

	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return &SDK{
		conf:   conf,
		client: client.NewYQClient(ctx, grpcConn),
	}, nil
}
