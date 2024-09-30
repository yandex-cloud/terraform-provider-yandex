resource "yandex_storage_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
    rule {
      default_retention {
        mode  = "GOVERNANCE"
        years = 1
      }
    }
  }
}
