---
subcategory: "Yandex Query"
page_title: "Yandex: yandex_yq_monitoring_connection"
description: |-
  Manages Object Storage connection.
---

# yandex_yq_monitoring_connection (Resource)

Manages Monitoring connection in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection).

## Example usage

```terraform
//
// Create a new Monitoring connection.
//

resource "yandex_yq_monitoring_connection" "my_mon_connection" {
  name               = "tf-test-mon-connection"
  description        = "Connection has been created from Terraform"
  folder_id          = "my_folder"
  service_account_id = yandex_iam_service_account.for-yq.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The resource name.

### Optional

- `description` (String) The resource description.
- `folder_id` (String) The folder identifier.
- `service_account_id` (String) The service account ID to access resources on behalf of.

### Read-Only

- `cloud_id` (String) The cloud identifier.
- `id` (String) The resource identifier.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud).

```shell
# terraform import yandex_yq_monitoring_connection.<resource Name> <resource Id>
terraform import yandex_yq_monitoring_connection.my_mon_connection ...
```
