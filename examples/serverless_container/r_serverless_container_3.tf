//
// Create a new Serverless Container with Storage mount.
//
locals {
  folder_id = "folder_id"
}

resource "yandex_serverless_container" "test-container-object-storage-mount" {
  name               = "some_name"
  memory             = 128
  service_account_id = yandex_iam_service_account.sa.id
  image {
    url = "cr.yandex/yc/test-image:v1"
  }
  mounts {
    mount_point_path = "/mount/point"
    mode             = "ro"
    object_storage {
      bucket = yandex_storage_bucket.my-bucket.bucket
    }
  }
}

// Auxiliary resources
resource "yandex_iam_service_account" "sa" {
  folder_id = local.folder_id
  name      = "test-sa"
}

resource "yandex_resourcemanager_folder_iam_member" "sa-editor" {
  folder_id = local.folder_id
  role      = "storage.editor"
  member    = "serviceAccount:${yandex_iam_service_account.sa.id}"
}

resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = yandex_iam_service_account.sa.id
  description        = "static access key for object storage"
}

resource "yandex_storage_bucket" "my-bucket" {
  access_key = yandex_iam_service_account_static_access_key.sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-static-key.secret_key
  bucket     = "bucket"
}
