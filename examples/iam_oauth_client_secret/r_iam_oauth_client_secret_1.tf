//
// Create a new OAuth Client Secret.
//
resource "yandex_iam_oauth_client_secret" "my-oauth-client-secret" {
  oauth_client_id = yandex_iam_oauth_client.my-oauth-client.id
  description     = "secret for oauth client"
  pgp_key         = "keybase:keybaseusername"
}

