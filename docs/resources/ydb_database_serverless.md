---
subcategory: "Managed Service for YDB"
page_title: "Yandex: yandex_ydb_database_serverless"
description: |-
  Manages Yandex Database serverless cluster.
---

# yandex_ydb_database_serverless (Resource)

Yandex Database (serverless) resource. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).

## Example usage

```terraform
//
// Create a new YDB Serverless Database.
//
resource "yandex_ydb_database_serverless" "my_ydb" {
  name      = "test-ydb-serverless"
  folder_id = data.yandex_resourcemanager_folder.test_folder.id

  deletion_protection = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the Yandex Database serverless cluster.

* `location_id` - (Optional) Location ID for the Yandex Database serverless cluster.

* `folder_id` - (Optional) ID of the folder that the Yandex Database serverless cluster belongs to. It will be deduced from provider configuration if not set explicitly.

* `description` - (Optional) A description for the Yandex Database serverless cluster.

* `labels` - (Optional) A set of key/value label pairs to assign to the Yandex Database serverless cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the database. Can be either `true` or `false`

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the Yandex Database serverless cluster.

* `document_api_endpoint` - Document API endpoint of the Yandex Database serverless cluster.

* `ydb_full_endpoint` - Full endpoint of the Yandex Database serverless cluster.

* `ydb_api_endpoint` - API endpoint of the Yandex Database serverless cluster. Useful for SDK configuration.

* `database_path` - Full database path of the Yandex Database serverless cluster. Useful for SDK configuration.

* `tls_enabled` - Whether TLS is enabled for the Yandex Database serverless cluster. Useful for SDK configuration.

* `created_at` - The Yandex Database serverless cluster creation timestamp.

* `status` - Status of the Yandex Database serverless cluster.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_ydb_database_serverless.<resource Name> <resource Id>
terraform import yandex_ydb_database_serverless.my_ydb ...
```
