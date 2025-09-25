//
// Get information about Trino access control by cluster ID.
//
data "yandex_trino_catalog" "trino_access_control" {
  cluster_id = yandex_trino_cluster.trino.id
}
