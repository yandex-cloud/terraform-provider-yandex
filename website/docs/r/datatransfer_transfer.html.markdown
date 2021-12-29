---
layout: "yandex"
page_title: "Yandex: yandex_datatransfer_transfer"
sidebar_current: "docs-yandex-datatransfer-transfer"
description: |-
  Manages a Data Transfer transfer within Yandex.Cloud.
---

# yandex\_datatransfer\_transfer

Manages a Data Transfer transfer. For more information, see [the official documentation](https://cloud.yandex.com/docs/data-transfer/).

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

resource "yandex_datatransfer_transfer" "pgpg_transfer" {
  folder_id = "some_folder_id"
  name = "pgpg"
  source_id = yandex_datatransfer_endpoint.pg_source.id
  target_id = yandex_datatransfer_endpoint.pg_target.id
  type = "SNAPSHOT_AND_INCREMENT"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the transfer.
* `type` - (Required) Type of the transfer. One of "SNAPSHOT_ONLY", "INCREMENT_ONLY", "SNAPSHOT_AND_INCREMENT".
* `source_id` - (Optional) ID of the source endpoint for the transfer.
* `target_id` - (Optional) ID of the target endpoint for the transfer.
* `description` - (Optional) Arbitrary description text for the transfer.
* `folder_id` - (Optional) ID of the folder to create the transfer in. If it is not provided, the default provider folder is used.
* `labels` - (Optional) A set of key/value label pairs to assign to the Data Transfer transfer.

## Attributes Reference

* `id` - (Computed) Identifier of a new Data Transfer transfer.
* `warning` - (Computed) Error description if transfer has any errors.

## Import

A transfer can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_datatransfer_transfer.foo transfer_id
```
