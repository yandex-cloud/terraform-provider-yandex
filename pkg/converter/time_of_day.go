package converter

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"google.golang.org/genproto/googleapis/type/timeofday"
)

// ParseTimeOfDay parse time of day string
func ParseTimeOfDay(ts string, diags *diag.Diagnostics) *timeofday.TimeOfDay {
	if ts == "" {
		return nil
	}
	t, err := time.Parse(time.Kitchen, ts)
	if err != nil {
		diags.AddError("Failed to parse time of day", fmt.Sprintf("Failed to parse time of day, got: %T, need like %s", ts, "4:00PM"))
		return nil
	}

	prototod := &timeofday.TimeOfDay{
		Hours:   int32(t.Hour()),
		Minutes: int32(t.Minute()),
		Seconds: int32(t.Second()),
		Nanos:   int32(t.Nanosecond()),
	}

	return prototod
}

func GetTimeOfDay(d1 *timeofday.TimeOfDay, stateValue string, diags *diag.Diagnostics) string {
	if stateValue == "" {
		date := time.Date(0, time.January, 1,
			int(d1.GetHours()), int(d1.GetMinutes()), int(d1.GetSeconds()), int(d1.GetNanos()), time.UTC)

		return date.Format(time.Kitchen)
	}
	t, err := time.Parse(time.Kitchen, stateValue)
	if err != nil {
		diags.AddError("Failed to parse time of day", fmt.Sprintf("Failed to parse time of day, got: %T, need like %s", stateValue, "4:00PM"))

		return ""
	}
	d2 := &timeofday.TimeOfDay{
		Hours:   int32(t.Hour()),
		Minutes: int32(t.Minute()),
		Seconds: int32(t.Second()),
		Nanos:   int32(t.Nanosecond()),
	}

	if d1 != nil && d2 != nil && d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos && d1.Hours == d2.Hours && d1.Minutes == d2.Minutes {
		return stateValue
	}

	date := time.Date(0, time.January, 1,
		int(d1.GetHours()), int(d1.GetMinutes()), int(d1.GetSeconds()), int(d1.GetNanos()), time.UTC)

	return date.Format(time.Kitchen)
}
