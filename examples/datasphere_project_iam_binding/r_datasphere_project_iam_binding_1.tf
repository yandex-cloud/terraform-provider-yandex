//
// Create a new Datasphere Project and new IAM Binding for it.
//
resource "yandex_datasphere_community" "my-community" {
  name               = "example-datasphere-community"
  description        = "Description of community"
  billing_account_id = "example-organization-id"
  organization_id    = "example-organization-id"
}

resource "yandex_datasphere_project" "my-project" {
  name        = "example-datasphere-project"
  description = "Datasphere Project description"

  community_id = yandex_datasphere_community.my-community.id
  # ...
}

resource "yandex_datasphere_project_iam_binding" "project-iam" {
  project_id = "your-datasphere-project-id"
  role       = "datasphere.community-projects.developer"
  members = [
    "system:allUsers",
  ]
}
