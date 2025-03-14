---
subcategory: "Managed Service for YDB"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Database serverless cluster.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Database serverless cluster. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).

## Example usage

{{ tffile "examples/ydb_database_serverless/d_ydb_database_serverless_1.tf" }}

## Argument Reference

The following arguments are supported:

* `database_id` - (Optional) ID of the Yandex Database serverless cluster.

* `name` - (Optional) Name of the Yandex Database serverless cluster.

* `folder_id` - (Optional) ID of the folder that the Yandex Database serverless cluster belongs to. It will be deduced from provider configuration if not set explicitly.

~> If `database_id` is not specified `name` and `folder_id` will be used to designate Yandex Database serverless cluster.

## Attributes Reference

* `location_id` - Location ID of the Yandex Database serverless cluster.

* `description` - A description of the Yandex Database serverless cluster.

* `labels` - A set of key/value label pairs assigned to the Yandex Database serverless cluster.

* `document_api_endpoint` - Document API endpoint of the Yandex Database serverless cluster.

* `ydb_full_endpoint` - Full endpoint of the Yandex Database serverless cluster.

* `ydb_api_endpoint` - API endpoint of the Yandex Database serverless cluster. Useful for SDK configuration.

* `database_path` - Full database path of the Yandex Database serverless cluster. Useful for SDK configuration.

* `tls_enabled` - Whether TLS is enabled for the Yandex Database serverless cluster. Useful for SDK configuration.

* `persistence_mode` - Persistence mode of the Yandex Database cluster. Useful for SDK configuration.

* `announce_hostnames` - Announce fqdn instead of ip address for the Yandex Database cluster. Useful for SDK configuration.

* `created_at` - The Yandex Database serverless cluster creation timestamp.

* `status` - Status of the Yandex Database serverless cluster.

* `deletion_protection` - Inhibits deletion of the database. Can be either `true` or `false`
