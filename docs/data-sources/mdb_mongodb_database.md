---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: yandex_mdb_mongodb_database"
description: |-
  Get information about a Yandex Managed MongoDB database.
---

# yandex_mdb_mongodb_database

Get information about a Yandex Managed MongoDB database. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/).

## Example usage

```terraform
data "yandex_mdb_mongodb_database" "foo" {
  cluster_id = "some_cluster_id"
  name       = "test"
}

output "owner" {
  value = data.yandex_mdb_mongodb_database.foo.name
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the MongoDB cluster.

* `name` - (Required) The name of the MongoDB cluster.
