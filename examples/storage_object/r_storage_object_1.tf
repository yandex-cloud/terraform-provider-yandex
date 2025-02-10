//
// Create a new Storage Object in Bucket.
//
resource "yandex_storage_object" "cute-cat-picture" {
  bucket = "cat-pictures"
  key    = "cute-cat"
  source = "/images/cats/cute-cat.jpg"
  tags = {
    test = "value"
  }
}
