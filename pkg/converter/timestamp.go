package converter

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ParseTimestamp ...
func ParseTimestamp(ts string, diags *diag.Diagnostics) *timestamppb.Timestamp {
	if ts == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		diags.AddError(
			"Failed to parse timestamp",
			fmt.Sprintf("Failed to parce timestamp, got: %T, need RFC3339 like %s", ts, time.RFC3339),
		)
		return nil
	}

	return timestamppb.New(t)
}
