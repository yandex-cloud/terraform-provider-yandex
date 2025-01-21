---
subcategory: "Managed Service for MongoDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Managed MongoDB database.
---


# {{.Name}}

{{ .Description }}


Get information about a Yandex Managed MongoDB database. For more information, see [the official documentation](https://cloud.yandex.com/docs/managed-mongodb/).

## Example usage

{{tffile "yandex-framework/docs-templates/mdb_mongodb_database/datasource-example-1.tf"}}

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the MongoDB cluster.

* `name` - (Required) The name of the MongoDB cluster.
