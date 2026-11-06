package yandex_organizationmanager_policy_authentication_policy_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/proto"

	policy "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/policy"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func preserveObligationDurations(ctx context.Context, _ *provider_config.Config, res *policy.AuthenticationPolicyRule,
	model, prior *yandexOrganizationmanagerPolicyAuthenticationPolicyRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if prior == nil || prior.Obligations.IsNull() || prior.Obligations.IsUnknown() || len(res.GetObligations()) == 0 {
		return diags
	}
	previous := prior.Obligations.Elements()
	obligations := make([]*policy.Obligation, len(previous))
	for i, value := range previous {
		raw, err := value.ToTerraformValue(ctx)
		if err != nil {
			diags.AddError("Failed to read prior obligation", err.Error())
			return diags
		}
		if raw.IsFullyKnown() {
			obligations[i] = expandYandexOrganizationmanagerPolicyAuthenticationPolicyRuleObligationStruct(ctx, value.(types.Object), &diags)
		}
	}
	if diags.HasError() {
		return diags
	}
	values := make([]attr.Value, 0, len(res.GetObligations()))
	for _, obligation := range res.GetObligations() {
		var value attr.Value
		for i, candidate := range obligations {
			if candidate != nil && proto.Equal(obligation, candidate) {
				value = previous[i]
				obligations[i] = nil
				break
			}
		}
		if value == nil {
			value = flattenYandexOrganizationmanagerPolicyAuthenticationPolicyRuleObligationStruct(ctx, obligation, NewYandexOrganizationmanagerPolicyAuthenticationPolicyRuleObligationStructModel(), &diags)
		}
		values = append(values, value)
	}
	if !diags.HasError() {
		diags.Append(tfsdk.ValueFrom(ctx, values, model.Obligations.Type(ctx), &model.Obligations)...)
	}
	return diags
}
