resource "yandex_datasphere_project_iam_binding" "project-iam" {
  project_id = "your-datasphere-project-id"
  role       = "datasphere.community-projects.developer"
  members = [
    "system:allUsers",
  ]
}
