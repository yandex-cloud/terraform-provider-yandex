resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  anonymous_access_flags {
    read        = true
    list        = false
    config_read = true
  }
}
