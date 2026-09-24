//
// Get information about an existing Managed OpenSearch user.
//
data "yandex_mdb_opensearch_user" "read_only" {
  cluster_id = "some_cluster_id"
  name       = "read_only"
}

output "connection_id" {
  value = data.yandex_mdb_opensearch_user.read_only.connection_manager.connection_id
}
