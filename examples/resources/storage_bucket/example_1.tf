provider "yandex" {
  zone = "ru-central1-a"
}

resource "yandex_storage_bucket" "test" {
  bucket = "tf-test-bucket"
}
