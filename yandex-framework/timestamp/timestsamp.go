package timestamp

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func Get(ts *timestamppb.Timestamp) string {
	const defaultTimeFormat = time.RFC3339

	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(defaultTimeFormat)
}
