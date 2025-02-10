---
subcategory: "Resource Manager"
page_title: "Yandex: yandex_resourcemanager_cloud"
description: |-
  Allows management of the Cloud resource.
---

# yandex_resourcemanager_cloud (Resource)

Allows creation and management of Cloud resources for an existing Yandex Cloud Organization. See [the official documentation](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy) for additional info. Note: deletion of clouds may take up to 30 minutes as it requires a lot of communication between cloud services.

## Example usage

```terraform
//
// Create a new Cloud.
//
resource "yandex_resourcemanager_cloud" "my_cloud" {
  organization_id = "my_organization_id"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Optional) Yandex Cloud Organization that the cloud belongs to. If value is omitted, the default provider Organization ID is used.

* `name` - (Optional) The name of the Cloud.

* `description` - (Optional) A description of the Cloud.

* `labels` - (Optional) A set of key/value label pairs to assign to the Cloud.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_resourcemanager_cloud.<resource Name> <resource Id>
terraform import yandex_resourcemanager_cloud.my_cloud b1g3a**********qc5df
```
