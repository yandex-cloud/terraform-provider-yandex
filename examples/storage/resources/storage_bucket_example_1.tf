locals {
  folder_id = "<folder-id>"
}

provider "yandex" {
  folder_id = local.folder_id
  zone      = "ru-central1-a"
}

resource "yandex_storage_bucket" "test" {
  bucket = "tf-test-bucket"
}
