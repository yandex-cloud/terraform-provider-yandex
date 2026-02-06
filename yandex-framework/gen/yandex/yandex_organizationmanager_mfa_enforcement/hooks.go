package yandex_organizationmanager_mfa_enforcement

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/genproto/protobuf/field_mask"
)

var mapCreateStatus = map[string]organizationmanager.CreateMfaEnforcementRequest_Status{
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_UNSPECIFIED.String(): organizationmanager.CreateMfaEnforcementRequest_STATUS_UNSPECIFIED,
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_ACTIVE.String():      organizationmanager.CreateMfaEnforcementRequest_STATUS_ACTIVE,
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_INACTIVE.String():    organizationmanager.CreateMfaEnforcementRequest_STATUS_INACTIVE,
}

var mapUpdateStatus = map[string]organizationmanager.UpdateMfaEnforcementRequest_Status{
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_UNSPECIFIED.String(): organizationmanager.UpdateMfaEnforcementRequest_STATUS_UNSPECIFIED,
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_ACTIVE.String():      organizationmanager.UpdateMfaEnforcementRequest_STATUS_ACTIVE,
	organizationmanager.MfaEnforcementStatus_MFA_ENFORCEMENT_STATUS_INACTIVE.String():    organizationmanager.UpdateMfaEnforcementRequest_STATUS_INACTIVE,
}

func setCorrectEnumForCreate(_ context.Context, _ *config.Config, req *organizationmanager.CreateMfaEnforcementRequest, plan *yandexOrganizationmanagerMfaEnforcementModel) diag.Diagnostics {
	req.SetStatus(mapCreateStatus[plan.Status.ValueString()])
	return nil
}

func setCorrectEnumForUpdate(_ context.Context, _ *config.Config, req *organizationmanager.UpdateMfaEnforcementRequest, plan, state *yandexOrganizationmanagerMfaEnforcementModel) diag.Diagnostics {
	if !plan.Status.Equal(state.Status) {
		if req.UpdateMask == nil {
			req.SetUpdateMask(&field_mask.FieldMask{Paths: []string{"status"}})
		}
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "status")
	}
	req.SetStatus(mapUpdateStatus[plan.Status.ValueString()])
	return nil
}
