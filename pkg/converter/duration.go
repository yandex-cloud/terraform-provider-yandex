package converter

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ParseDuration parse duration string into durationpb
func ParseDuration(ts string, diags *diag.Diagnostics) *durationpb.Duration {
	if ts == "" {
		return nil
	}
	t, err := time.ParseDuration(ts)
	if err != nil {
		diags.AddError(
			"Failed to parse duration",
			fmt.Sprintf("Failed to parce duration, got: %T, need like %s", ts, "1h10m10s"),
		)
		return nil
	}

	return durationpb.New(t)
}

func GetDuration(d1 *durationpb.Duration, stateValue string, diags *diag.Diagnostics) string {
	if stateValue == "" {
		return d1.AsDuration().String()
	}
	t, err := time.ParseDuration(stateValue)
	if err != nil {
		diags.AddError(
			"Failed to parse duration",
			fmt.Sprintf("Failed to parce duration, got: %T, need like %s", stateValue, "1h10m10s"),
		)
		return ""
	}
	d2 := durationpb.New(t)

	if d1 != nil && d2 != nil && d1.Seconds == d2.Seconds && d1.Nanos == d2.Nanos {
		return stateValue
	}

	return d1.AsDuration().String()
}
