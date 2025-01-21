package auth

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func PrepareUpdateRequest(ctx context.Context, clusterID string, plan *model.AuthSettings) (*opensearch.UpdateAuthSettingsRequest, diag.Diagnostics) {
	req := &opensearch.UpdateAuthSettingsRequest{
		ClusterId: clusterID,
	}

	if plan == nil {
		return req, diag.Diagnostics{}
	}

	if plan.SAML.IsNull() {
		req.Settings = &opensearch.AuthSettings{
			Saml: nil,
		}

		return req, diag.Diagnostics{}
	}

	saml, diags := model.SAMLFromState(ctx, plan.SAML)
	if diags.HasError() {
		return nil, diags
	}

	idpMetadataFile := saml.IdpMetadataFileContent.ValueString()

	samlSettings := &opensearch.SAMLSettings{
		Enabled:         saml.Enabled.ValueBool(),
		IdpEntityId:     saml.IdpEntityID.ValueString(),
		IdpMetadataFile: []byte(idpMetadataFile),
		SpEntityId:      saml.SpEntityID.ValueString(),
		DashboardsUrl:   saml.DashboardsUrl.ValueString(),
		RolesKey:        saml.RolesKey.ValueString(),
		SubjectKey:      saml.SubjectKey.ValueString(),
	}

	req.Settings = &opensearch.AuthSettings{
		Saml: samlSettings,
	}

	return req, diag.Diagnostics{}
}
