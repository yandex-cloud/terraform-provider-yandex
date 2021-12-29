---
layout: "yandex"
page_title: "Yandex: yandex_datatransfer_endpoint"
sidebar_current: "docs-yandex-datatransfer-endpoint"
description: |-
  Manages a Data Transfer endpoint within Yandex.Cloud.
---

# yandex\_datatransfer\_endpoint

Manages a Data Transfer endpoint. For more information, see [the official documentation](https://cloud.yandex.com/docs/data-transfer/).

## Example Usage

```hcl
resource "yandex_datatransfer_endpoint" "pg_source" {
  name = "pg-test-source"
  settings {
    postgres_source {
      connection {
        on_premise {
          hosts = [
            "example.org"
          ]
          port = 5432
        }
      }
      slot_gigabyte_lag_limit = 100
      database = "db1"
      user = "user1"
      password {
        raw = "123"
      }
    }
  }
}

resource "yandex_datatransfer_endpoint" "pg_target" {
  folder_id = "some_folder_id"
  name = "pg-test-target2"
  settings {
    postgres_target {
      connection {
        mdb_cluster_id = "some_cluster_id"
      }
      database = "db2"
      user = "user2"
      password {
        raw = "321"
      }
    }
  }
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the endpoint.
* `settings` - (Required) Settings for the endpoint. The structure is documented below.
* `description` - (Optional) Arbitrary description text for the endpoint.
* `folder_id` - (Optional) ID of the folder to create the endpoint in. If it is not provided, the default provider folder is used.
* `labels` - (Optional) A set of key/value label pairs to assign to the Data Transfer endpoint.

The `settings` block supports:

* `postgres_source` - (Optional) Settings specific to the PostgreSQL source endpoint.
* `postgres_target` - (Optional) Settings specific to the PostgreSQL target endpoint.
* `mysql_source` - (Optional) Settings specific to the MySQL source endpoint.
* `mysql_target` - (Optional) Settings specific to the MySQL target endpoint.

For the documentation of the specific endpoint settings see below.

---

The `postgres_source` block supports:

* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `service_schema` - (Optional) Name of the database schema in which auxiliary tables needed for the transfer will be created. Empty `service_schema` implies schema "public".
* `include_tables` - List of tables to transfer, formatted as `schemaname.tablename`. If omitted or an empty list is specified, all tables will be transferred.
* `exclude_tables` - List of tables which will not be transfered, formatted as `schemaname.tablename`.
* `object_transfer_settings` - (Optional) Defines which database schema objects should be transferred, e.g. views, functions, etc.
* `slot_gigabyte_lag_limit` - (Optional) Maximum WAL size held by the replication slot, in gigabytes. Exceeding this limit will result in a replication failure and deletion of the replication slot. Unlimited by default.

The `postgres_target` block supports:

* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.

The `connection` block supports exactly one of the following attributes:
* `mdb_cluster_id` - Identifier of the Managed PostgreSQL cluster.
* `on_premise` - Connection settings of the on-premise PostgreSQL server.

The `on_premise` block supports:
* `hosts` - (Required) List of host names of the PostgreSQL server. Exactly one host is expected currently.
* `port` - (Required) Port for the database connection.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `object_transfer_settings` block supports the following attributes:
* `sequence`
* `sequence_owned_by`
* `table`
* `primary_key`
* `fk_constraint`
* `default_values`
* `constraint`
* `index`
* `view`
* `function`
* `trigger`
* `type`
* `rule`
* `collation`
* `policy`
* `cast`
All of the attrubutes are optional and should be either "BEFORE_DATA", "AFTER_DATA" or "NEVER".

---

The `mysql_source` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `include_tables_regex` - (Optional) List of regular expressions of table names which should be transferred. A table name is formatted as schemaname.tablename. For example, a single regular expression may look like `^mydb.employees$`.
* `exclude_tables_regex` - (Optional) Opposite of `include_table_regex`. The tables matching the specified regular expressions will not be transferred.
* `timezone` - (Optional) Timezone to use for parsing timestamps for saving source timezones. Accepts values from IANA timezone database. Default: local timezone.
* `object_transfer_settings` - (Optional) Defines which database schema objects should be transferred, e.g. views, routines, etc.

The `mysql_target` block supports:
* `connection` - (Required) Connection settings. The structure is documented below.
* `database` - (Required) Name of the database to transfer.
* `user` - (Required) User for the database access.
* `password` - (Required) Password for the database access. This is a block with a single field named `raw` which should contain the password.
* `sql_mode` - (Optional) [sql_mode](https://dev.mysql.com/doc/refman/5.7/en/sql-mode.html) to use when interacting with the server. Defaults to "NO_AUTO_VALUE_ON_ZERO,NO_DIR_IN_CREATE,NO_ENGINE_SUBSTITUTION".
* `skip_constraint_checks` - (Optional) When true, disables foreign key checks. See [foreign_key_checks](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_foreign_key_checks). False by default.
* `timezone` - (Optional) Timezone to use for parsing timestamps for saving source timezones. Accepts values from IANA timezone database. Default: local timezone.

The `connection` block supports exactly one of the following attributes:
* `mdb_cluster_id` - Identifier of the Managed MySQL cluster.
* `on_premise` - Connection settings of the on-premise MySQL server.

The `on_premise` block supports:
* `hosts` - (Required) List of host names of the MySQL server. Exactly one host is expected currently.
* `port` - (Required) Port for the database connection.
* `tls_mode` - (Optional) TLS settings for the server connection. Empty implies plaintext connection. The structure is documented below.
* `subnet_id` - (Optional) Identifier of the Yandex Cloud VPC subnetwork to user for accessing the database. If omitted, the server has to be accessible via Internet.

The `object_transfer_settings` block supports the following attributes:
* `view`
* `routine`
* `trigger`
All of the attrubutes are optional and should be either "BEFORE_DATA", "AFTER_DATA" or "NEVER".

---

The `tls_mode` block supports exactly one of the following attributes:
* `disabled` - Empty block designating that the connection is not secured, i.e. plaintext connection.
* `enabled` - If this attribute is not an empty block, then TLS is used for the server connection. The structure is documented below.

The `enabled` block supports:
* `ca_certificate` - (Optional) X.509 certificate of the certificate authority which issued the server's certificate, in PEM format. If empty, the server's certificate must be signed by a well-known CA.

## Attributes Reference

* `id` - (Computed) Identifier of a new Data Transfer endpoint.
* `created_at` - (Computed) Data Transfer endpoint creation timestamp.
* `author` - (Computed) Identifier of the IAM user account of the user who created the endpoint.

## Import

An endpoint can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_datatransfer_endpoint.foo endpoint_id
```
