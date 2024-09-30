resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  https {
    certificate_id = "<certificate_id_from_certificate_manager>"
  }
}
