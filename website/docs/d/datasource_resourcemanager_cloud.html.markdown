---
layout: "yandex"
page_title: "Yandex: yandex_resourcemanager_cloud"
sidebar_current: "docs-yandex-datasource-resourcemanager-cloud"
description: |-
  Retrieve Yandex RM Cloud details.
---

# yandex\_resourcemanager\_cloud

Use this data source to get cloud details.
For more information, see [Cloud](https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#cloud).

## Example Usage

```hcl
data "yandex_resourcemanager_cloud" "foo" {
  name = "foo-cloud"
}

output "cloud_create_timestamp" {
  value = "${data.yandex_resourcemanager_cloud.foo.created_at}"
}
```

## Argument Reference

The following arguments are supported:

* `cloud_id` - (Optional) ID of the cloud.
* `name` - (Optional) Name of the cloud.

~> **NOTE:** Either `cloud_id` or `name` must be specified.

## Attributes Reference

The following attributes are returned:

* `name` - Name of the cloud.
* `description` - Description of the cloud.
* `created_at` - Cloud creation timestamp.
* `folders` - List of folders in the cloud
