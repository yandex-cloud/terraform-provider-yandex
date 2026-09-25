package topicreadercommon

import (
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/clone"
)

type PublicReadSelector struct {
	Path       string
	Partitions []int64
	ReadFrom   time.Time     // zero value mean skip read from filter
	MaxTimeLag time.Duration // 0 mean skip time lag filter
}

// Clone create deep clone of the selector
func (s PublicReadSelector) Clone() *PublicReadSelector { //nolint:gocritic
	s.Partitions = clone.Int64Slice(s.Partitions)

	return &s
}
