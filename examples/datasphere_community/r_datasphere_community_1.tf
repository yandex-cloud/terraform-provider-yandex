//
// Create a new Datasphere Community.
//
resource "yandex_datasphere_community" "my-community" {
  name               = "example-datasphere-community"
  description        = "Description of community"
  billing_account_id = "example-organization-id"
  labels = {
    "foo" : "bar"
  }
  organization_id = "example-organization-id"
}
