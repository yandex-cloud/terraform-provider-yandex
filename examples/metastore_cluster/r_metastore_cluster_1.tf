//
// Create a new Metastore Cluster.
//
resource "yandex_metastore_cluster" "my_metastore_cluster" {
  name               = "metastore-created-with-terraform"
  subnet_ids         = [yandex_vpc_subnet.a.id]
  security_group_ids = [yandex_vpc_security_group.metastore-sg.id]
  service_account_id = yandex_iam_service_account.for-metastore.id

  cluster_config = {
    resource_preset_id = "c2-m8"
  }

  maintenance_window = {
    type = "WEEKLY"
    day  = "MON"
    hour = 12
  }

  description = "My awesome Metastore"

  logging = {
    enabled   = true
    folder_id = var.folder_id
    min_level = "INFO"
  }
}
