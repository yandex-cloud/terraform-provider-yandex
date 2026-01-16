package customplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

func CloudStoragePlanModifier() planmodifier.Object {
	return &cloudStoragePlanModifierStruct{}
}

type cloudStoragePlanModifierStruct struct{}

func (m *cloudStoragePlanModifierStruct) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	var planModel models.CloudStorage
	diags := req.PlanValue.As(ctx, &planModel, datasize.DefaultOpts)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if planModel.Enabled.IsNull() || planModel.Enabled.IsUnknown() || planModel.Enabled.ValueBool() {
		return
	}

	// Cloud storage disabled
	planModel.MoveFactor = types.NumberNull()
	planModel.DataCacheEnabled = types.BoolNull()
	planModel.DataCacheMaxSize = types.Int64Null()
	planModel.PreferNotToMerge = types.BoolNull()

	newPlanVal, diags := types.ObjectValueFrom(ctx, models.CloudStorageAttrTypes, planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.PlanValue = newPlanVal
}

func (m *cloudStoragePlanModifierStruct) Description(context.Context) string {
	return `
		Cloud storage block plan modifier. 
		Reset all fields when enabled equal false.
	`
}

func (m *cloudStoragePlanModifierStruct) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
