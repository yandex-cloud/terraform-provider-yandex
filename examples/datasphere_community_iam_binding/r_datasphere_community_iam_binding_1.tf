//
// Create a new Datasphere Community and new IAM Binding for it.
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

resource "yandex_datasphere_community_iam_binding" "community-iam" {
  community_id = yandex_datasphere_community.my-community.id
  role         = "datasphere.communities.developer"
  members = [
    "system:allUsers",
  ]
}
