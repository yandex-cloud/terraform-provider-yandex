package changefreeze

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	maintenance "github.com/yandex-cloud/go-genproto/yandex/cloud/maintenance/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestMain registers the flags used by Terraform sweeper runs.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestIgnoreTerminateError(t *testing.T) {
	pastPeriodStatus, err := status.New(codes.FailedPrecondition, "past period").WithDetails(&maintenance.ChangeFreezeErrorCode{
		Type: maintenance.ChangeFreezeErrorCode_CANNOT_DELETE_PAST_PERIOD,
	})
	if err != nil {
		t.Fatalf("attach change freeze error detail: %v", err)
	}
	invalidTimeRangeStatus, err := status.New(codes.InvalidArgument, "invalid range").WithDetails(&maintenance.ChangeFreezeErrorCode{
		Type: maintenance.ChangeFreezeErrorCode_INVALID_TIME_RANGE,
	})
	if err != nil {
		t.Fatalf("attach change freeze error detail: %v", err)
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil},
		{name: "plain error", err: fmt.Errorf("plain error")},
		{name: "not found", err: status.Error(codes.NotFound, "not found"), want: true},
		{name: "past period", err: pastPeriodStatus.Err(), want: true},
		{name: "wrapped past period", err: fmt.Errorf("operation failed: %w", pastPeriodStatus.Err()), want: true},
		{name: "other typed error", err: invalidTimeRangeStatus.Err()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IgnoreTerminateError(test.err); got != test.want {
				t.Fatalf("IgnoreTerminateError() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestParseImportID(t *testing.T) {
	resourceID, changeFreezeID, err := parseImportID(" cluster-id : freeze-id ")
	if err != nil {
		t.Fatalf("parse valid import ID: %v", err)
	}
	if resourceID != "cluster-id" || changeFreezeID != "freeze-id" {
		t.Fatalf("unexpected parsed IDs: %q, %q", resourceID, changeFreezeID)
	}

	for _, id := range []string{"", "cluster-id", ":freeze-id", "cluster-id:", "a:b:c"} {
		if _, _, err := parseImportID(id); err == nil {
			t.Fatalf("expected %q to be rejected", id)
		}
	}
}

func TestIsActiveReplacement(t *testing.T) {
	unchangedState := []types.String{types.StringValue("resource"), types.StringValue("start"), types.StringValue("end"), types.StringNull()}
	changedPlan := append([]types.String(nil), unchangedState...)
	changedPlan[1] = types.StringValue("new-start")
	unknownPlan := append([]types.String(nil), unchangedState...)
	unknownPlan[2] = types.StringUnknown()

	tests := []struct {
		name    string
		status  types.String
		current []types.String
		planned []types.String
		want    bool
	}{
		{name: "active unchanged", status: types.StringValue("ACTIVE"), current: unchangedState, planned: unchangedState},
		{name: "active changed", status: types.StringValue("ACTIVE"), current: unchangedState, planned: changedPlan, want: true},
		{name: "active unknown replacement value", status: types.StringValue("ACTIVE"), current: unchangedState, planned: unknownPlan, want: true},
		{name: "scheduled changed", status: types.StringValue("SCHEDULED"), current: unchangedState, planned: changedPlan},
		{name: "completed changed", status: types.StringValue("COMPLETED"), current: unchangedState, planned: changedPlan},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isActiveReplacement(test.status, test.current, test.planned); got != test.want {
				t.Fatalf("isActiveReplacement() = %v, want %v", got, test.want)
			}
		})
	}
}
