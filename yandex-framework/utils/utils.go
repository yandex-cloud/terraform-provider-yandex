package utils

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const DefaultTimeout = 1 * time.Minute
const DefaultPageSize = 1000
const defaultTimeFormat = time.RFC3339

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetTimestamp(timestamp *timestamppb.Timestamp) string {
	if timestamp == nil {
		return ""
	}
	return timestamp.AsTime().Format(defaultTimeFormat)
}
