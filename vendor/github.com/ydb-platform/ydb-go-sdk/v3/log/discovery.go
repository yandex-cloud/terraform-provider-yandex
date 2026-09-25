package log

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/kv"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Discovery makes trace.Discovery with logging events from details
func Discovery(l Logger, d trace.Detailer, opts ...Option) (t trace.Discovery) {
	return internalDiscovery(wrapLogger(l, opts...), d)
}

func internalDiscovery(l Logger, d trace.Detailer) (t trace.Discovery) {
	t.OnDiscover = func(info trace.DiscoveryDiscoverStartInfo) func(trace.DiscoveryDiscoverDoneInfo) {
		if d.Details()&trace.DiscoveryEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, DEBUG, "ydb", "discovery", "list", "endpoints")
		l.Log(ctx, "discovery starting...",
			kv.String("address", info.Address),
			kv.String("database", info.Database),
		)
		start := time.Now()

		return func(info trace.DiscoveryDiscoverDoneInfo) {
			if info.Error == nil {
				l.Log(WithLevel(ctx, INFO), "discovery done",
					kv.Latency(start),
					kv.Stringer("endpoints", kv.Endpoints(info.Endpoints)),
				)
			} else {
				l.Log(WithLevel(ctx, ERROR), "discovery failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}
	t.OnWhoAmI = func(info trace.DiscoveryWhoAmIStartInfo) func(doneInfo trace.DiscoveryWhoAmIDoneInfo) {
		if d.Details()&trace.DiscoveryEvents == 0 {
			return nil
		}
		ctx := with(*info.Context, TRACE, "ydb", "discovery", "whoAmI")
		l.Log(ctx, "discovery whoami starting...")
		start := time.Now()

		return func(info trace.DiscoveryWhoAmIDoneInfo) {
			if info.Error == nil {
				l.Log(ctx, "discovery whoami done",
					kv.Latency(start),
					kv.String("user", info.User),
					kv.Strings("groups", info.Groups),
				)
			} else {
				l.Log(WithLevel(ctx, WARN), "discovery whoami failed",
					kv.Error(info.Error),
					kv.Latency(start),
					kv.Version(),
				)
			}
		}
	}

	return t
}
