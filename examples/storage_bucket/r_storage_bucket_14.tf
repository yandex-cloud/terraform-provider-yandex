//
// Set Bucket Folder Id.
//
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  folder_id = "<folder_id>"
}
