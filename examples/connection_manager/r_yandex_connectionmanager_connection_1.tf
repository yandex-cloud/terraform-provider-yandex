resource "yandex_connectionmanager_connection" "test-connection" {
  folder_id = "folder_id"

  name = "my_connection"
  description = "my_connection description"

  labels = {
    "key" = "value"
  }

  params = {
    postgresql = {
      managed_cluster_id = "cluster_id"
      auth = {
        user_password = {
          user = "name"
          password = {
            raw: "password"
          }
        }
      }
    }
  }
}
