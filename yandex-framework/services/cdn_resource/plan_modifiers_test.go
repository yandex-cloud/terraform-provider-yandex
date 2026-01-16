package cdn_resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestUseUnknownOnUpdate_Create(t *testing.T) {
	// During create (state is null), the modifier should not change the plan value
	ctx := context.Background()

	// Create request with null state (create operation)
	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{}, nil), // null state
		},
		Plan: tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{}),
		},
		PlanValue: types.StringUnknown(),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier := UseUnknownOnUpdate()
	modifier.PlanModifyString(ctx, req, resp)

	// Plan value should remain unknown (not modified)
	if !resp.PlanValue.IsUnknown() {
		t.Errorf("Expected plan value to remain unknown during create, got: %v", resp.PlanValue)
	}
}

func TestUseUnknownOnUpdate_Update(t *testing.T) {
	// During update (state exists), the modifier should set plan value to unknown
	ctx := context.Background()

	// Create request with existing state (update operation)
	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"updated_at": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"updated_at": tftypes.NewValue(tftypes.String, "2024-01-01T00:00:00Z"),
			}),
		},
		Plan: tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"updated_at": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"updated_at": tftypes.NewValue(tftypes.String, "2024-01-01T00:00:00Z"),
			}),
		},
		StateValue: types.StringValue("2024-01-01T00:00:00Z"),
		PlanValue:  types.StringValue("2024-01-01T00:00:00Z"),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier := UseUnknownOnUpdate()
	modifier.PlanModifyString(ctx, req, resp)

	// Plan value should be set to unknown during update
	if !resp.PlanValue.IsUnknown() {
		t.Errorf("Expected plan value to be unknown during update, got: %v", resp.PlanValue)
	}
}

func TestUseUnknownOnUpdate_Delete(t *testing.T) {
	// During delete (plan is null), the modifier should not change anything
	ctx := context.Background()

	// Create request with null plan (delete operation)
	req := planmodifier.StringRequest{
		State: tfsdk.State{
			Raw: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"updated_at": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"updated_at": tftypes.NewValue(tftypes.String, "2024-01-01T00:00:00Z"),
			}),
		},
		Plan: tfsdk.Plan{
			Raw: tftypes.NewValue(tftypes.Object{}, nil), // null plan = delete
		},
		StateValue: types.StringValue("2024-01-01T00:00:00Z"),
		PlanValue:  types.StringNull(),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}

	modifier := UseUnknownOnUpdate()
	modifier.PlanModifyString(ctx, req, resp)

	// Plan value should remain null during delete
	if !resp.PlanValue.IsNull() {
		t.Errorf("Expected plan value to remain null during delete, got: %v", resp.PlanValue)
	}
}

func TestUseUnknownOnUpdate_Description(t *testing.T) {
	modifier := UseUnknownOnUpdate()

	desc := modifier.Description(context.Background())
	if desc == "" {
		t.Error("Expected non-empty description")
	}

	mdDesc := modifier.MarkdownDescription(context.Background())
	if mdDesc == "" {
		t.Error("Expected non-empty markdown description")
	}
}
