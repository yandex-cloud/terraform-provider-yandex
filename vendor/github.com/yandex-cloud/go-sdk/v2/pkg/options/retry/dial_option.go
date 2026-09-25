package retry

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
)

// RetryDialOption returns DialOption that configures the default
// service config, which will be used in cases where:
//
// 1. User provide service config in SDK build after RetryDialOption function.
// In this case retries configuration override user's service config.
//
// 2. User don't get service config through xDS. In that case service config
// from xDS override retries configuration.
//
// It's recommended to use default configuration (DefaultRetryDialOption)
// for retries in SDK to avoid retry amplification.
func RetryDialOption(opts ...RetryOption) (grpc.DialOption, error) {
	config := defaultRetryConfig()
	for _, opt := range opts {
		opt(config)
	}

	if config.mc == nil {
		return nil, fmt.Errorf("can't provide retry config with this options")
	}

	mc := []*methodConfig{}
	for _, v := range config.mc {
		mc = append(mc, v)
	}

	c, err := json.Marshal(&grpcRetryPolicy{MethodConfig: mc, RetryThrottling: *config.retryThrottling})
	if err != nil {
		return nil, err
	}

	return grpc.WithDefaultServiceConfig(string(c)), nil
}

// DefaultRetryDialOption returns DialOption that configures the default
// service config, which will be used in cases where:
//
// 1. User provide service config in SDK build after DefaultRetryDialOption function.
// In this case retries configuration override user's service config.
//
// 2. User don't get service config through xDS. In that case service config
// from xDS override retries configuration.
func DefaultRetryDialOption() (grpc.DialOption, error) {
	return RetryDialOption(WithDefaultRetryConfig())
}
