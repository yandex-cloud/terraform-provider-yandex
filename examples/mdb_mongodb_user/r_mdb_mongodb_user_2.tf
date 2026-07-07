//
// Create a new MDB MongoDB user authenticated via IAM.
//
// An IAM user is identified by the ID of an IAM subject (for example, a
// service account) and authenticates with IAM tokens, so it has no password.
//
resource "yandex_iam_service_account" "my_sa" {
  name = "mongodb-iam-user"
}

resource "yandex_mdb_mongodb_user" "my_iam_user" {
  cluster_id = yandex_mdb_mongodb_cluster.my_cluster.id
  name       = yandex_iam_service_account.my_sa.id
  auth_type  = "IAM"

  permission {
    database_name = "db1"
    roles         = ["readWrite"]
  }
}
