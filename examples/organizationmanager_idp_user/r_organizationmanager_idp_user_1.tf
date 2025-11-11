//
// Create a new OrganizationManager Idp User.
//
resource "yandex_organizationmanager_idp_user" "example_user" {
  userpool_id = yandex_organizationmanager_idp_userpool.your_userpool.userpool_id
  username    = "example@your-domain.com"
  full_name   = "Test User"
  given_name  = "Test"
  family_name = "User"
  email       = "test-userov@example.com"
  is_active   = true
  password_spec = {
    password = "secret-password"
  }
}
