//
// Get information about existing MDB PostgreSQL Backup Retention Policy.
//
data "yandex_mdb_postgresql_backup_retention_policy" "my_policy" {
  cluster_id = "some_cluster_id"
  policy_id  = "some_policy_id"
}

output "policy_name" {
  value = data.yandex_mdb_postgresql_backup_retention_policy.my_policy.policy_name
}
