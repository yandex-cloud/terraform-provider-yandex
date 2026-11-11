package organizationmanager_idp_application_oauth_application_assignment

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type assignmentModel struct {
	ApplicationId types.String `tfsdk:"application_id"`
	SubjectId     types.String `tfsdk:"subject_id"`
}
