package api

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type adminPasswordModifier struct{}

func (m adminPasswordModifier) Description(_ context.Context) string {
	return "Special handling of admin_password attribute."
}

func (m adminPasswordModifier) MarkdownDescription(_ context.Context) string {
	return "Special handling of admin_password attribute."
}

func (m adminPasswordModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.ValueString() == AdminPasswordStubOnImport {
		if !req.ConfigValue.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				req.Path,
				"The actual value of the admin_password used by Apache Airflow may differ from the one specified in the configuration",
				"This warning occurs because the Airflow cluster resource has been imported."+
					" When importing, it is not possible to get the value of the admin password set when creating the cluster."+
					" Therefore, terraform state does not contain this information and terraform cannot detect its changes."+
					" To eliminate this warning, you can either:\n\n"+
					" * remove the admin_password attribute from the cluster configuration\n\n"+
					" * or manually save the admin_password value in state, which will match the value set in the configuration\n",
			)
		}
		resp.PlanValue = req.StateValue
	} else {
		if req.ConfigValue.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required argument",
				"The argument \"admin_password\" is required, but no definition was found.",
			)
			return
		}

		resp.RequiresReplace = true
	}
}
