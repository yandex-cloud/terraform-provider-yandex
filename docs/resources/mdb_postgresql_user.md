---
subcategory: "Managed Service for PostgreSQL"
page_title: "Yandex: yandex_mdb_postgresql_user"
description: |-
  Manages a PostgreSQL user within Yandex Cloud.
---

# yandex_mdb_postgresql_user (Resource)

Manages a PostgreSQL user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).

## Example usage

```terraform
//
// Create a new MDB PostgreSQL database User.
//
resource "yandex_mdb_postgresql_user" "my_user" {
  cluster_id = yandex_mdb_postgresql_cluster.my_cluster.id
  name       = "alice"
  password   = "password"
  conn_limit = 50
  settings = {
    default_transaction_isolation = "read committed"
    log_min_duration_statement    = 5000
  }
}

resource "yandex_mdb_postgresql_cluster" "my_cluster" {
  name        = "test"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id

  config {
    version = 15
    resources {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 16
    }
  }

  host {
    zone      = "ru-central1-d"
    subnet_id = yandex_vpc_subnet.foo.id
  }
}

// Auxiliary resources
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_id` (String) The ID of the PostgreSQL cluster.
- `name` (String) The name of the user.

### Optional

- `conn_limit` (Number) The maximum number of connections per user. (Default 50).
- `deletion_protection` (String) The `true` value means that resource is protected from accidental deletion.
- `generate_password` (Boolean) Generate password using Connection Manager. Allowed values: true or false. It's used only during user creation and is ignored during updating.

~> **Must specify either password or generate_password**.
- `grants` (List of String) List of the user's grants.
- `login` (Boolean) User's ability to login.
- `password` (String, Sensitive) The password of the user.
- `permission` (Block Set) Set of permissions granted to the user. (see [below for nested schema](#nestedblock--permission))
- `settings` (Map of String) Map of user settings. [Full description](https://yandex.cloud/docs/managed-postgresql/api-ref/grpc/Cluster/create#yandex.cloud.mdb.postgresql.v1.UserSettings).

* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. One of:  - 0: `unspecified`
  - 1: `read uncommitted`
  - 2: `read committed`
  - 3: `repeatable read`
  - 4: `serializable`

* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)

* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)

* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way. One of:
  - 0: `unspecified`
  - 1: `on`
  - 2: `off`
  - 3: `local`
  - 4: `remote write`
  - 5: `remote apply`

* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.

* `log_statement` - This setting specifies which SQL statements should be logged (on the user level). One of:
  - 0: `unspecified`
  - 1: `none`
  - 2: `ddl`
  - 3: `mod`
  - 4: `all`

* `pool_mode` - Mode that the connection pooler is working in with specified user. One of:
  - 1: `session`
  - 2: `transaction`
  - 3: `statement`

* `prepared_statements_pooling` - This setting allows user to use prepared statements with transaction pooling. Boolean.

* `catchup_timeout` - The connection pooler setting. It determines the maximum allowed replication lag (in seconds). Pooler will reject connections to the replica with a lag above this threshold. Default value is 0, which disables this feature. Integer.

* `wal_sender_timeout` - The maximum time (in milliseconds) to wait for WAL replication (can be set only for PostgreSQL 12+). Terminate replication connections that are inactive for longer than this amount of time. Integer.

* `idle_in_transaction_session_timeout` - Sets the maximum allowed idle time (in milliseconds) between queries, when in a transaction. Value of 0 (default) disables the timeout. Integer.

* `statement_timeout` - The maximum time (in milliseconds) to wait for statement. Value of 0 (default) disables the timeout. Integer
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `connection_manager` (Map of String) Connection Manager connection configuration. Filled in by the server automatically.
- `id` (String) The ID of this resource.

<a id="nestedblock--permission"></a>
### Nested Schema for `permission`

Required:

- `database_name` (String) The name of the database that the permission grants access to.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_mdb_postgresql_user.<resource Name> <resource Id>
terraform import yandex_mdb_postgresql_user.my_user ...
```
