package organizationmanager_mfa_enforcement_audience

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type audienceModel struct {
	MfaEnforcementId types.String `tfsdk:"mfa_enforcement_id"`
	SubjectId        types.String `tfsdk:"subject_id"`
}
