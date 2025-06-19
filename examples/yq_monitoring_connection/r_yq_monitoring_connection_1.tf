//
// Create a new Monitoring connection.
//

resource "yandex_yq_monitoring_connection" "my_mon_connection" {
  name               = "tf-test-mon-connection"
  description        = "Connection has been created from Terraform"
  folder_id          = "my_folder"
  service_account_id = yandex_iam_service_account.for-yq.id
}
