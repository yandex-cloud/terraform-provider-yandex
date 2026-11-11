package changefreeze

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	maintenance "github.com/yandex-cloud/go-genproto/yandex/cloud/maintenance/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var replacementAttributePaths = []path.Path{
	path.Root("resource_id"),
	path.Root("start_at"),
	path.Root("end_at"),
	path.Root("reason"),
}

// IgnoreTerminateError reports whether Terminate has already reached the
// outcome expected by Terraform deletion.
func IgnoreTerminateError(err error) bool {
	if err == nil {
		return false
	}
	grpcStatus, ok := status.FromError(err)
	if !ok || grpcStatus == nil {
		return false
	}
	if grpcStatus.Code() == codes.NotFound {
		return true
	}

	for _, detail := range grpcStatus.Details() {
		changeFreezeError, ok := detail.(*maintenance.ChangeFreezeErrorCode)
		if ok && changeFreezeError.GetType() == maintenance.ChangeFreezeErrorCode_CANNOT_DELETE_PAST_PERIOD {
			return true
		}
	}

	return false
}

// ImportState imports a change freeze using the resource_id:change_freeze_id
// composite identifier. The API does not return resource_id, so it must be
// retained in Terraform state explicitly.
func ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resourceID, changeFreezeID, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_id"), resourceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("change_freeze_id"), changeFreezeID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), changeFreezeID)...)
}

func parseImportID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected import ID format: resource_id:change_freeze_id, got: %s", id)
	}

	resourceID := strings.TrimSpace(parts[0])
	changeFreezeID := strings.TrimSpace(parts[1])
	if resourceID == "" || changeFreezeID == "" {
		return "", "", fmt.Errorf("both resource_id and change_freeze_id must be non-empty")
	}

	return resourceID, changeFreezeID, nil
}

// PreventActiveReplacement blocks a destroy-and-create replacement of an
// active change freeze. A plain destroy is deliberately allowed: Delete will
// terminate the active period without attempting to create another one.
func PreventActiveReplacement(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var currentStatus types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("status"), &currentStatus)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentValues := make([]types.String, 0, len(replacementAttributePaths))
	plannedValues := make([]types.String, 0, len(replacementAttributePaths))
	for _, attributePath := range replacementAttributePaths {
		var currentValue types.String
		var plannedValue types.String
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, attributePath, &currentValue)...)
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, attributePath, &plannedValue)...)
		if resp.Diagnostics.HasError() {
			return
		}
		currentValues = append(currentValues, currentValue)
		plannedValues = append(plannedValues, plannedValue)
	}

	if isActiveReplacement(currentStatus, currentValues, plannedValues) {
		resp.Diagnostics.AddError(
			"Cannot replace an active change freeze",
			"Terraform cannot replace a change freeze while it is ACTIVE. Replacement first terminates the active period, after which service overlap and minimum-gap constraints can prevent creation of the new period. Wait until the period is completed, or remove it from the configuration to terminate it without creating a replacement.",
		)
	}
}

func isActiveReplacement(currentStatus types.String, currentValues, plannedValues []types.String) bool {
	if currentStatus.IsNull() || currentStatus.IsUnknown() || currentStatus.ValueString() != maintenance.ChangeFreeze_ACTIVE.String() {
		return false
	}
	if len(currentValues) != len(plannedValues) {
		return true
	}

	for i := range currentValues {
		if !currentValues[i].Equal(plannedValues[i]) {
			return true
		}
	}

	return false
}
