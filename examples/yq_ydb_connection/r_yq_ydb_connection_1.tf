//
// Create a new YDB connection.
//

resource "yandex_yq_ydb_connection" "my_ydb_connection" {
    name = "tf-test-ydb-connection"
    description = "Connection has been created from Terraform"
    database_id = "db_id"
    service_account_id = yandex_iam_service_account.for-yq.id
}
