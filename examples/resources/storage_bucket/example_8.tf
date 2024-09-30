resource "yandex_storage_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"
}

resource "yandex_storage_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  logging {
    target_bucket = yandex_storage_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
