//
// Create a new Yandex Cloud Function with mounted Object Storage Bucket.
//
resource "yandex_function" "test-function" {
  name               = "some_name"
  user_hash          = "v1"
  runtime            = "python37"
  entrypoint         = "index.handler"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = yandex_iam_service_account.sa.id
  content {
    zip_filename = "function.zip"
  }
  mounts {
    name = "mnt"
    mode = "ro"
    object_storage {
      bucket = yandex_storage_bucket.my-bucket.bucket
    }
  }
}

locals {
  folder_id = "folder_id"
}

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
