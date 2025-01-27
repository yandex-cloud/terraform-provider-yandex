//
// Configure the Yandex Cloud Provider (Advanced)
//
provider "yandex_second" {
  token                    = "auth_token_here"
  service_account_key_file = "path_to_service_account_key_file"
  cloud_id                 = "cloud_id_here"
  folder_id                = "folder_id_here"
  zone                     = "ru-central1-d"
  shared_credentials_file  = "path_to_shared_credentials_file"
  profile                  = "testing"
}
