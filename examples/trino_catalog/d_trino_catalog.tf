//
// Get information about Trino catalog by name
//
data "yandex_trino_catalog" "trino_catalog_by_name" {
  cluster_id = yandex_trino_cluster.trino.id
  name       = "catalog"
}

//
// Get information about Trino catalog by id
//
data "yandex_trino_catalog" "trino_catalog_by_id" {
  cluster_id = yandex_trino_cluster.trino.id
  id         = "<tirno-catalog-id>"
}
