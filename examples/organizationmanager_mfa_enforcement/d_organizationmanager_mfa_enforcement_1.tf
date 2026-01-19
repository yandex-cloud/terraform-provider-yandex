//
// OrganizationManager MFA Enforcement.
//
data "yandex_organizationmanager_mfa_enforcement" "example_mfa_enforcement" {
  acr_id 	 	      = "any-mfa"
}