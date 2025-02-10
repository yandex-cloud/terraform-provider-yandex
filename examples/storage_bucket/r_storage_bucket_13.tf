//
// Set Bucket Max Size.
//
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  max_size = 1048576
}
