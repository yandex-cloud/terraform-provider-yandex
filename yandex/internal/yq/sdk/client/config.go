package client

import "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/client/connection"

type YQConfig struct {
	grpcEndpoint string
	dialer       connection.ConnectionDialer
}

func NewYQConfig(grpcEndpoint string, dialer connection.ConnectionDialer) *YQConfig {
	return &YQConfig{
		grpcEndpoint: grpcEndpoint,
		dialer:       dialer,
	}
}
