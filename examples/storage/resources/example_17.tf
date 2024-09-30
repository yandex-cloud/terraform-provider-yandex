resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  default_storage_class = "COLD"
}
