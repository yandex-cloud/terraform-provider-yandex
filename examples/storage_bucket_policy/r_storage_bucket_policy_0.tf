provider "yandex" {
  cloud_id           = "<my_cloud_id>"
  folder_id          = "<my_folder_id>"
  storage_access_key = "<my_storage_access_key>"
  storage_secret_key = "<my_storage_secret_key>"
  token              = "<my_iam_token>"
}

resource "yandex_storage_bucket_policy" "my_policy_0" {
  bucket = "my_bucket_name_0"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::my-policy-bucket/*",
        "arn:aws:s3:::my-policy-bucket"
      ]
    },
    {
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": [
        "arn:aws:s3:::my-policy-bucket/*",
        "arn:aws:s3:::my-policy-bucket"
      ]
    }
  ]
}
POLICY
}
