package datasphere_community

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

// convertToTerraformModel Convert from the Proto community data model to the Terraform community data model
// and refresh any attribute values.
func convertToTerraformModel(ctx context.Context, terraformModel *communityDataModel, grpcModel *datasphere.Community, diag *diag.Diagnostics) {
	terraformModel.Name = types.StringValue(grpcModel.Name)
	terraformModel.CreatedAt = types.StringValue(timestamp.Get(grpcModel.CreatedAt))
	terraformModel.Description = types.StringValue(grpcModel.Description)
	terraformModel.CreatedBy = types.StringValue(grpcModel.CreatedById)
	terraformModel.OrganizationId = types.StringValue(grpcModel.OrganizationId)

	labels, diags := types.MapValueFrom(ctx, types.StringType, grpcModel.Labels)
	terraformModel.Labels = labels
	diag.Append(diags...)
}
