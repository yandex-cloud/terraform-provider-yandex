//
// Create a new KMS Assymetric Signature Key and new IAM Binding for it.
//
resource "yandex_kms_asymmetric_signature_key" "your-key" {
  name = "asymmetric-signature-key-name"
}

resource "yandex_kms_asymmetric_signature_key_iam_binding" "viewer" {
  asymmetric_signaturen_key_id = yandex_kms_asymmetric_signature_key.your-key.id
  role                         = "viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
