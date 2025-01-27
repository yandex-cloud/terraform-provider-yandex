resource "yandex_ydb_database_serverless" "database1" {
  name      = "test-ydb-serverless"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id
}

resource "yandex_ydb_database_iam_binding" "viewer" {
  database_id = yandex_ydb_database_serverless.database1.id
  role        = "ydb.viewer"

  members = [
    "userAccount:foo_user_id",
  ]
}
