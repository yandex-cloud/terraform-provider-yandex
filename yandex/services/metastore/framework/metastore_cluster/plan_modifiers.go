package metastore_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func mwPlanModifier() planmodifier.Object {
	return &maintenanceWindowPlanModifier{}
}

type maintenanceWindowPlanModifier struct{}

func (m *maintenanceWindowPlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.PlanValue.IsNull() && !req.PlanValue.IsUnknown() {
		return
	}

	mw := MaintenanceWindowValue{
		MaintenanceWindowType: types.StringValue("ANYTIME"),
		state:                 attr.ValueStateKnown,
	}

	mwObj, diags := mw.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	resp.PlanValue = mwObj
}

func (m *maintenanceWindowPlanModifier) Description(context.Context) string {
	return `
		Maintenance window block plan modifier. 
		Sets maintenance window to ANYTIME if it is not specified in plan or has null value.
	`
}

func (m *maintenanceWindowPlanModifier) MarkdownDescription(context.Context) string {
	return `
		Maintenance window block plan modifier. 
		Sets maintenance window to *ANYTIME* if it was not specified in plan or has null value.
	`
}
