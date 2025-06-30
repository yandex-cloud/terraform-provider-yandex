//
// Create a new Object Storage (S3) Bucket IAM Binding.
//
resource "yandex_storage_bucket_iam_binding" "bucket-iam" {
  bucket = "your-bucket-name"
  role      = "storage.admin"

  members = [
    "system:allUsers",
  ]
}
