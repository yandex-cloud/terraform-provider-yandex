//
// Create a new Trino catalog
//

resource "yandex_trino_catalog" "catalog" {
  name        = "name"
  description = "descriptionr"
  cluster_id  = yandex_trino_cluster.trino.id
  postgresql = {
    connection_manager = {
      connection_id = "<connection_id>"
      database      = "database-name"
      connection_properties = {
        "targetServerType" = "primary"
      }
    }
  }
}
