package cdn_resource

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/go-sdk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// getShieldingLocation retrieves the current shielding location for a CDN resource
// Returns nil if shielding is not enabled (NotFound error is expected in this case)
func getShieldingLocation(ctx context.Context, resourceID string, sdk *ycsdk.SDK) (*int64, error) {
	resp, err := sdk.CDN().Shielding().Get(ctx, &cdn.GetShieldingDetailsRequest{
		ResourceId: resourceID,
	})

	// NotFound is not an error - it means shielding is disabled
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shielding details: %w", err)
	}

	return &resp.LocationId, nil
}

// updateShieldingIfChanged updates shielding configuration if it has changed between plan and state
// This is called from Update method
func updateShieldingIfChanged(ctx context.Context, plan, state *CDNResourceModel, sdk *ycsdk.SDK) error {
	// Check if shielding has changed
	if plan.Shielding.Equal(state.Shielding) {
		return nil
	}

	resourceID := state.ID.ValueString()

	// Case 1: Enable shielding (plan has value, state is null/empty)
	if !plan.Shielding.IsNull() && plan.Shielding.ValueString() != "" {
		locationID, err := strconv.ParseInt(plan.Shielding.ValueString(), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid shielding location ID: %w", err)
		}
		return enableShielding(ctx, resourceID, locationID, sdk)
	}

	// Case 2: Disable shielding (plan is null/empty, state has value)
	return disableShielding(ctx, resourceID, sdk)
}

// applyShieldingFromPlan applies shielding configuration from plan (used in Create)
func applyShieldingFromPlan(ctx context.Context, plan *CDNResourceModel, sdk *ycsdk.SDK) error {
	if plan.Shielding.IsNull() || plan.Shielding.ValueString() == "" {
		// Shielding not specified in plan - nothing to do
		return nil
	}

	locationID, err := strconv.ParseInt(plan.Shielding.ValueString(), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid shielding location ID: %w", err)
	}

	return enableShielding(ctx, plan.ID.ValueString(), locationID, sdk)
}

// enableShielding activates shielding for a CDN resource at the specified location
func enableShielding(ctx context.Context, resourceID string, locationID int64, sdk *ycsdk.SDK) error {
	op, err := sdk.WrapOperation(
		sdk.CDN().Shielding().Activate(ctx, &cdn.ActivateShieldingRequest{
			ResourceId: resourceID,
			LocationId: locationID,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to activate shielding: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for shielding activation: %w", err)
	}

	return nil
}

// disableShielding deactivates shielding for a CDN resource
func disableShielding(ctx context.Context, resourceID string, sdk *ycsdk.SDK) error {

	currentShielding, err := getShieldingLocation(ctx, resourceID, sdk)
	if err != nil {
		return fmt.Errorf("failed to get current shielding status: %w", err)
	}

	// If shielding is already disabled (nil), do nothing
	if currentShielding == nil {
		return nil
	}
	op, err := sdk.WrapOperation(
		sdk.CDN().Shielding().Deactivate(ctx, &cdn.DeactivateShieldingRequest{
			ResourceId: resourceID,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to deactivate shielding: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for shielding deactivation: %w", err)
	}

	return nil
}

// flattenShielding converts shielding location ID to Terraform string value
func flattenShielding(shielding *int64) types.String {
	if shielding == nil {
		return types.StringNull()
	}
	return types.StringValue(fmt.Sprintf("%d", *shielding))
}
