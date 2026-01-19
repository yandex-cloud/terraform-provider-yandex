//
// Create a new OrganizationManager MFA Enforcement.
//
resource "yandex_organizationmanager_mfa_enforcement" "example_mfa_enforcement" {
  name            = "example-mfa-enforcement"
  organization_id = "your_organization_id"
  acr_id 	 	      = "any-mfa"
  ttl 			      = "2h45m"
  status 		      = "MFA_ENFORCEMENT_STATUS_ACTIVE"
  enroll_window   = "2h45m"
  description     = "Description example"
}
