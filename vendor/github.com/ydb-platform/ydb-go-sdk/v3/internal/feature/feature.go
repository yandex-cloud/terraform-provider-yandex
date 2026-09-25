package feature

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

type Flag int

const (
	Unknown Flag = iota
	Enabled
	Disabled
)

func (f Flag) ToYDB() Ydb.FeatureFlag_Status {
	switch f {
	case Enabled:
		return Ydb.FeatureFlag_ENABLED
	case Disabled:
		return Ydb.FeatureFlag_DISABLED
	case Unknown:
		return Ydb.FeatureFlag_STATUS_UNSPECIFIED
	default:
		panic("ydb: unknown feature flag status")
	}
}

func FromYDB(f Ydb.FeatureFlag_Status) Flag {
	switch f {
	case Ydb.FeatureFlag_ENABLED:
		return Enabled
	case Ydb.FeatureFlag_DISABLED:
		return Disabled
	case Ydb.FeatureFlag_STATUS_UNSPECIFIED:
		return Unknown
	default:
		panic("ydb: unknown Ydb feature flag status")
	}
}
