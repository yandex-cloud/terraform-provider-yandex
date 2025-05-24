//
// Create a new YDS connection.
//

resource "yandex_yq_yds_connection" "my_yds_connection" {
    name = "tf-test-yds-connection"
    description = "Connection has been created from Terraform"
    database_id = "db_id"
    service_account_id = yandex_iam_service_account.for-yq.id
}
