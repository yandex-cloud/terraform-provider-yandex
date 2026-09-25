package coordination

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Coordination"

type ConsistencyMode uint

const (
	ConsistencyModeUnset ConsistencyMode = iota
	ConsistencyModeStrict
	ConsistencyModeRelaxed

	consistencyAggregated = "Aggregated"
	consistencyDetailed   = "Detailed"
	consistencyRelaxed    = "Relaxed"
	consistencyStrict     = "Strict"
	consistencyUnknown    = "Unknown"
	consistencyUnset      = "Unset"
)

func (t ConsistencyMode) String() string {
	switch t {
	default:
		return consistencyUnknown
	case ConsistencyModeUnset:
		return consistencyUnset
	case ConsistencyModeStrict:
		return consistencyStrict
	case ConsistencyModeRelaxed:
		return consistencyRelaxed
	}
}

type RatelimiterCountersMode uint

const (
	RatelimiterCountersModeUnset RatelimiterCountersMode = iota
	RatelimiterCountersModeAggregated
	RatelimiterCountersModeDetailed
)

func (t RatelimiterCountersMode) String() string {
	switch t {
	default:
		return consistencyUnknown
	case RatelimiterCountersModeUnset:
		return consistencyUnset
	case RatelimiterCountersModeAggregated:
		return consistencyAggregated
	case RatelimiterCountersModeDetailed:
		return consistencyDetailed
	}
}

type NodeConfig struct {
	Path                     string
	SelfCheckPeriodMillis    uint32
	SessionGracePeriodMillis uint32
	ReadConsistencyMode      ConsistencyMode
	AttachConsistencyMode    ConsistencyMode
	RatelimiterCountersMode  RatelimiterCountersMode
}

func (t ConsistencyMode) To() Ydb_Coordination.ConsistencyMode {
	switch t {
	case ConsistencyModeStrict:
		return Ydb_Coordination.ConsistencyMode_CONSISTENCY_MODE_STRICT
	case ConsistencyModeRelaxed:
		return Ydb_Coordination.ConsistencyMode_CONSISTENCY_MODE_RELAXED
	default:
		return Ydb_Coordination.ConsistencyMode_CONSISTENCY_MODE_UNSET
	}
}

func (t RatelimiterCountersMode) To() Ydb_Coordination.RateLimiterCountersMode {
	switch t {
	case RatelimiterCountersModeAggregated:
		return Ydb_Coordination.RateLimiterCountersMode_RATE_LIMITER_COUNTERS_MODE_AGGREGATED
	case RatelimiterCountersModeDetailed:
		return Ydb_Coordination.RateLimiterCountersMode_RATE_LIMITER_COUNTERS_MODE_DETAILED
	default:
		return Ydb_Coordination.RateLimiterCountersMode_RATE_LIMITER_COUNTERS_MODE_UNSET
	}
}
