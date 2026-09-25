package conn

import (
	"time"

	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type Config interface {
	DialTimeout() time.Duration
	ConnectionTTL() time.Duration
	Trace() *trace.Driver
	GrpcDialOptions() []grpc.DialOption
}
