---
subcategory: "Managed Service for OpenSearch"
---

# yandex_mdb_opensearch_user (DataSource)

Retrieves information about an OpenSearch user, including the ID of its managed Connection Manager connection.

## Example usage

```terraform
data "yandex_mdb_opensearch_user" "read_only" {
  cluster_id = "some_cluster_id"
  name       = "read_only"
}

data "yandex_connectionmanager_connection" "read_only_connection" {
  connection_id = data.yandex_mdb_opensearch_user.read_only.connection_manager.connection_id
}

output "connection_user" {
  value = data.yandex_connectionmanager_connection.read_only_connection.params.opensearch.auth.user_password.user
}
```

The `connection_manager.connection_id` attribute identifies the managed Connection Manager connection associated with the OpenSearch user. Pass it to the [`yandex_connectionmanager_connection`](connectionmanager_connection.md) data source to retrieve the connection parameters and user information.

## Arguments & Attributes Reference

- `cluster_id` (**Required**)(String). ID of the OpenSearch cluster.
- `name` (**Required**)(String). Name of the OpenSearch user.
- `connection_manager` (*Read-Only*) (Attributes). Connection Manager connection associated with the user.
  - `connection_id` (String). ID of the Connection Manager connection.
- `timeouts` (Attributes). Timeout settings.
  - `read` (String). Timeout for read operations. Default is `30m`.
