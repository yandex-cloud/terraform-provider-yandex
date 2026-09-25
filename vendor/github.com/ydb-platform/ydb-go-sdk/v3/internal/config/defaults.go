package config

import (
	"time"

	"google.golang.org/grpc/keepalive"
)

var (
	// DefaultKeepaliveInterval contains default duration between grpc keepalive
	DefaultKeepaliveInterval    = 10 * time.Second
	MinKeepaliveInterval        = 10 * time.Second
	DefaultDialTimeout          = 5 * time.Second
	DefaultGRPCMsgSize          = 64 * 1024 * 1024 // 64MB
	DefaultGrpcConnectionPolicy = keepalive.ClientParameters{
		Time:                DefaultKeepaliveInterval,
		Timeout:             MinKeepaliveInterval,
		PermitWithoutStream: true,
	}
)
