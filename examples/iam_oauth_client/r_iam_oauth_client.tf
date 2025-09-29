//
// Create a new OAuthClient.
//
resource "yandex_iam_oauth_client" "my-oauth-client" {
  name          = "my-oauth-client"
  folder_id 	= "aje5a**********qspd3"
  redirect_uris = ["https://localhost"]
  scopes        = ["iam"]
}
