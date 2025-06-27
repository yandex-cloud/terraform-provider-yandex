provider "yandex" {
  cloud_id           = "<my_cloud_id>"
  folder_id          = "<my_folder_id>"
  storage_access_key = "<my_storage_access_key>"
  storage_secret_key = "<my_storage_secret_key>"
  token              = "<my_iam_token>"
}

resource "yandex_storage_bucket_grant" "my_grant_0" {
  bucket = "my_bucket_name_0"
  grant {
    id          = "<user_id>"
    permissions = ["READ", "WRITE", "FULL_CONTROL"]
    type        = "CanonicalUser"
  }
}

resource "yandex_storage_bucket_grant" "my_grant_1" {
  bucket = "my_bucket_name_1"
  grant {
    id          = "<user_id>"
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
  grant {
    uri         = "<group_uri>"
    permissions = ["READ"]
    type        = "Group"
  }
}
