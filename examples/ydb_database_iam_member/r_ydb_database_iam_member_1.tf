//
// Create a new YDB Serverless Database and grant one IAM member access to it.
//
resource "yandex_ydb_database_serverless" "database1" {
  name      = "test-ydb-serverless"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id
}

resource "yandex_ydb_database_iam_member" "viewer" {
  database_id = yandex_ydb_database_serverless.database1.id
  role        = "ydb.viewer"
  member      = "userAccount:foo_user_id"
}
