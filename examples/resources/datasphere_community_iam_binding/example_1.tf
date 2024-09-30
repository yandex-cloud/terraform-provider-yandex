resource "yandex_datasphere_community_iam_binding" "community-iam" {
  community_id = "your-datasphere-community-id"
  role         = "datasphere.communities.developer"
  members = [
    "system:allUsers",
  ]
}
