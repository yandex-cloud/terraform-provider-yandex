---
layout: "yandex"
page_title: "Yandex: yandex_mdb_postgresql_user"
sidebar_current: "docs-yandex-datasource-mdb-postgresql-user"
description: |-
  Get information about a Yandex Managed PostgreSQL user.
---

# yandex\_mdb\_postgresql\_user

Get information about a Yandex Managed PostgreSQL user. For more information, see
[the official documentation](https://cloud.yandex.com/docs/managed-postgresql/).

## Example Usage

```hcl
data "yandex_mdb_postgresql_user" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "permission" {
  value = "${data.yandex_mdb_postgresql_user.foo.permission}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the PostgreSQL cluster.

* `name` - (Required) The name of the PostgreSQL user.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are
exported:

* `password` - The password of the user.
* `permission` - Set of permissions granted to the user. The structure is documented below.
* `login` - User's ability to login.
* `grants` - List of the user's grants.
* `conn_limit` - The maximum number of connections per user.
* `settings` - Map of user settings.
* `deletion_protection` - Inhibits deletion of the user.

The `permission` block supports:

* `database_name` - The name of the database that the permission grants access to.

The `settings` block supports:
Full description https://cloud.yandex.com/en-ru/docs/managed-postgresql/api-ref/grpc/user_service#UserSettings1

* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. One of:
  - 0: "unspecified"
  - 1: "read uncommitted"
  - 2: "read committed"
  - 3: "repeatable read"
  - 4: "serializable"

* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)

* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)

* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way. One of:
  - 0: "unspecified"
  - 1: "on"
  - 2: "off"
  - 3: "local"
  - 4: "remote write"
  - 5: "remote apply"

* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.

* `log_statement` - This setting specifies which SQL statements should be logged (on the user level). One of:
  - 0: "unspecified"
  - 1: "none"
  - 2: "ddl"
  - 3: "mod"
  - 4: "all"

* `pool_mode` - Mode that the connection pooler is working in with specified user. One of:
  - 0: "session"
  - 1: "transaction"
  - 2: "statement"

* `prepared_statements_pooling` - This setting allows user to use prepared statements with transaction pooling. Boolean.

* `catchup_timeout` - The connection pooler setting. It determines the maximum allowed replication lag (in seconds). Pooler will reject connections to the replica with a lag above this threshold. Default value is 0, which disables this feature. Integer.

* `wal_sender_timeout` - The maximum time (in milliseconds) to wait for WAL replication (can be set only for PostgreSQL 12+). Terminate replication connections that are inactive for longer than this amount of time. Integer.

* `idle_in_transaction_session_timeout` - Sets the maximum allowed idle time (in milliseconds) between queries, when in a transaction. Value of 0 (default) disables the timeout. Integer.

* `statement_timeout` - The maximum time (in milliseconds) to wait for statement. Value of 0 (default) disables the timeout. Integer