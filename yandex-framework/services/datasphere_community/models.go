package datasphere_community

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type communityDataModel struct {
	Id               types.String   `tfsdk:"id"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	CreatedBy        types.String   `tfsdk:"created_by"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	Labels           types.Map      `tfsdk:"labels"`
	OrganizationId   types.String   `tfsdk:"organization_id"`
	BillingAccountId types.String   `tfsdk:"billing_account_id"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
