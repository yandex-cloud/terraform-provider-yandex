package value

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

const InfiniteDuration = time.Duration(math.MaxInt64)

const (
	secondsPerMinute          uint64 = 60
	secondsPerHour                   = 60 * secondsPerMinute
	secondsPerDay                    = 24 * secondsPerHour
	microsecondsPerSecond            = 1e6
	nanosecondsPerMicrosecond        = 1000
)

// Date format layouts described in time.Format and time.ANSIC docs.
const (
	LayoutDate        = "2006-01-02"
	LayoutDatetime    = "2006-01-02T15:04:05Z"
	LayoutTimestamp   = "2006-01-02T15:04:05.000000Z"
	LayoutTzDate      = "2006-01-02"
	LayoutTzDatetime  = "2006-01-02T15:04:05"
	LayoutTzTimestamp = "2006-01-02T15:04:05.000000"
)

var epoch = time.Unix(0, 0)

// IntervalToDuration returns time.Duration from given microseconds
func IntervalToDuration(n int64) time.Duration {
	return time.Duration(n) * time.Microsecond
}

// Interval64ToDuration returns time.Duration from given nanoseconds
func Interval64ToDuration(n int64) time.Duration {
	return time.Duration(n) * time.Nanosecond
}

// durationToMicroseconds returns microseconds from given time.Duration
func durationToMicroseconds(d time.Duration) int64 {
	return int64(d / time.Microsecond)
}

// durationToNanoseconds returns nanoseconds from given time.Duration
func durationToNanoseconds(d time.Duration) int64 {
	return int64(d / time.Nanosecond)
}

// DateToTime up to 11761191-01-20 00:00:00 +0000 UTC.
func DateToTime(n uint32) time.Time {
	return time.Unix(0, 0).Add(time.Hour * 24 * time.Duration(n))
}

// Date32ToTime up to 11761191-01-20 00:00:00 +0000 UTC.
func Date32ToTime(days int32) time.Time {
	return time.Unix(int64(days)*24*60*60, 0)
}

// DatetimeToTime converts seconds to time.Time
// Up to 2106-02-07 06:28:15 +0000 UTC.
func DatetimeToTime(n uint32) time.Time {
	return time.Unix(int64(n), 0)
}

// Datetime64ToTime converts seconds to time.Time
// Up to 2106-02-07 06:28:15 +0000 UTC.
func Datetime64ToTime(n int64) time.Time {
	return time.Unix(n, 0)
}

// TimestampToTime converts given microseconds to time.Time
// Up to 586524-01-19 08:01:49.000551615 +0000 UTC.
func TimestampToTime(n uint64) time.Time {
	sec := n / microsecondsPerSecond
	nsec := (n - (sec * microsecondsPerSecond)) * nanosecondsPerMicrosecond

	return time.Unix(int64(sec), int64(nsec))
}

// Timestamp64ToTime converts given microseconds to time.Time
// Up to 586524-01-19 08:01:49.000551615 +0000 UTC.
func Timestamp64ToTime(n int64) time.Time {
	sec := n / microsecondsPerSecond
	nsec := (n - (sec * microsecondsPerSecond)) * nanosecondsPerMicrosecond

	return time.Unix(sec, nsec)
}

func TzDateToTime(s string) (t time.Time, err error) {
	ss := strings.Split(s, ",")
	if len(ss) != 2 { //nolint:mnd
		return t, xerrors.WithStackTrace(fmt.Errorf("not found timezone location part in '%s'", s))
	}
	location, err := time.LoadLocation(ss[1])
	if err != nil {
		return t, xerrors.WithStackTrace(err)
	}
	t, err = time.ParseInLocation(LayoutDate, ss[0], location)
	if err != nil {
		return t, xerrors.WithStackTrace(fmt.Errorf("parse '%s' failed: %w", s, err))
	}

	return t, nil
}

func TzDatetimeToTime(s string) (t time.Time, err error) {
	ss := strings.Split(s, ",")
	if len(ss) != 2 { //nolint:mnd
		return t, xerrors.WithStackTrace(fmt.Errorf("not found timezone location part in '%s'", s))
	}
	location, err := time.LoadLocation(ss[1])
	if err != nil {
		return t, xerrors.WithStackTrace(err)
	}
	t, err = time.ParseInLocation(LayoutTzDatetime, ss[0], location)
	if err != nil {
		return t, xerrors.WithStackTrace(fmt.Errorf("parse '%s' failed: %w", s, err))
	}

	return t, nil
}

func TzTimestampToTime(s string) (t time.Time, err error) {
	ss := strings.Split(s, ",")
	if len(ss) != 2 { //nolint:mnd
		return t, xerrors.WithStackTrace(fmt.Errorf("not found timezone location part in '%s'", s))
	}
	location, err := time.LoadLocation(ss[1])
	if err != nil {
		return t, xerrors.WithStackTrace(err)
	}
	layout := LayoutTzTimestamp
	if strings.IndexByte(ss[0], '.') < 0 {
		layout = LayoutTzDatetime
	}
	t, err = time.ParseInLocation(layout, ss[0], location)
	if err != nil {
		return t, xerrors.WithStackTrace(fmt.Errorf("parse '%s' failed: %w", s, err))
	}

	return t, nil
}
