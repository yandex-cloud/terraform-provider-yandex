resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  tags = {
    test_key  = "test_value"
    other_key = "other_value"
  }
}
