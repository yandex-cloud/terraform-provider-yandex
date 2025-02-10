//
// Get information about existing MDB MongoDB Database.
//
data "yandex_mdb_mongodb_database" "my_db" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "owner" {
  value = data.yandex_mdb_mongodb_database.my_db.name
}
