resource "yandex_trino_access_control" "trino_access_control" {
  cluster_id  = yandex_trino_cluster.trino.id
  catalogs = [
    {
      catalog = {
        ids = [
          yandex_trino_catalog.iceberg.id,
          yandex_trino_catalog.postgres.id,
        ]
      }
      users       = ["<iam_user_id>"]
      groups      = ["<iam_group_id>"]
      description = "Catalog access rule"
      permission  = "ALL"
    },
    {
      catalog = {
        name_regexp = "prod_.*"
      }
      permission = "NONE"
    },
    {
      permission = "READ_ONLY"
    },
  ]
}
