//
// Create new grants on an existing Storage Bucket.
//
resource "yandex_storage_bucket_grant" "my_bucket_grant" {
  bucket = "my_bucket_name_0"
  grant {
    id          = "<user0_id>"
    permissions = ["READ", "WRITE"]
    type        = "CanonicalUser"
  }
  grant {
    id          = "<user1_id>"
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
    permissions = ["READ"]
    type        = "Group"
  }
}
