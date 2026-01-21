package yandex_organizationmanager_idp_userpool_domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// customDomainImporter implements custom import logic for composite ID format: userpool_id:domain
func customDomainImporter(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: userpool_id:domain, got: %s", req.ID),
		)
		return
	}

	userpoolId := strings.TrimSpace(parts[0])
	domain := strings.TrimSpace(parts[1])

	if userpoolId == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"userpool_id cannot be empty",
		)
		return
	}

	if domain == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"domain cannot be empty",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("userpool_id"), userpoolId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), userpoolId+":"+domain)...)
}

// setID sets the composite ID in the model based on userpool_id and domain
func setID(ctx context.Context, providerConfig *provider_config.Config, res *idp.Domain, state *yandexOrganizationmanagerIdpUserpoolDomainModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if state.UserpoolId.IsNull() || state.UserpoolId.IsUnknown() {
		diags.AddError(
			"Failed to set ID",
			"userpool_id is required to set composite ID",
		)
		return diags
	}

	if state.Domain.IsNull() || state.Domain.IsUnknown() {
		diags.AddError(
			"Failed to set ID",
			"domain is required to set composite ID",
		)
		return diags
	}

	compositeID := state.UserpoolId.ValueString() + ":" + state.Domain.ValueString()
	state.Id = types.StringValue(compositeID)

	return diags
}
