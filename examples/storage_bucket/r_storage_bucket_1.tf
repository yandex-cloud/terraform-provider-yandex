//
// Create a new Storage Bucket. 
//
provider "yandex" {
  zone = "ru-central1-a"
}

resource "yandex_storage_bucket" "test_bucket" {
  bucket = "tf-test-bucket"
}
