package sdk

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client/connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/object_storage_connection"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// SDK is Yandex Query SDK
type SDK struct {
	conf       Config
	connDialer connection.ConnectionDialer
}

func NewYQSDK(ctx context.Context, conf Config, opts ...grpc.DialOption) (*SDK, error) {
	if conf.FolderID == "" {
		return nil, fmt.Errorf("\"folder_id\" is required to YQ SDK")
	}
	if conf.AuthToken == "" {
		return nil, fmt.Errorf("\"token\" is required to YQ SDK")
	}
	if conf.Endpoint == "" {
		return nil, fmt.Errorf("\"yq_endpoint\" is required to YQ SDK")
	}

	mdMiddleware := newYQMDMiddleware(conf.AuthToken, conf.FolderID)

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

	return &SDK{
		conf:       conf,
		connDialer: connection.NewSimpleConnectionDialer(dialOpts...),
	}, nil
}

func (s *SDK) ObjectStorageConnectionCaller(ctx context.Context) (object_storage_connection.ObjectStorageClient, error) {
	conf := client.NewYQConfig(s.conf.Endpoint, s.connDialer)

	yqClient, err := client.NewYQClient(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("YQ SDK new client: %w", err)
	}

	return yqClient, nil
}
