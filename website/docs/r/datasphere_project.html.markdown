---
layout: "yandex"
page_title: "Yandex: yandex_datasphere_project"
sidebar_current: "docs-yandex-datasphere-project"
description: |-
  Allows management of a Yandex.Cloud Datasphere project.
---

# yandex\_datasphere\_project

Allows management of Yandex Cloud Datasphere Communities

## Example Usage

```hcl
resource "yandex_datasphere_project" "my-project" {
  name = "example-datasphere-project"
  description = "Datasphere Project description"

  labels = {
    "foo": "bar"
  }

  community_id = yandex_datasphere_community.my-community.id

  limits = {
    max_units_per_hour = 10
    max_units_per_execution = 10
    balance = 10
  }

  settings = {
    service_account_id = yandex_iam_service_account.my-account.id
    subnet_id = yandex_vpc_subnet.my-subnet.id
    commit_mode = "AUTO"
    data_proc_cluster_id = "foo-data-proc-cluster-id"
    security_group_ids = [yandex_vpc_security_group.my-security-group.id]
    ide = "JUPYTER_LAB"
    default_folder_id = "foo-folder-id"
    stale_exec_timeout_mode = "ONE_HOUR"
  }
}
```

## Argument Reference

The following arguments are supported:

* `community_id` - (Required) Community ID where project would be created
* `name` - (Required) Name of the Datasphere Project.
* `description` - (Optional) Datasphere project description.
* `labels` - (Optional) A set of key/value label pairs to assign to the Datasphere Project.
* `limits` - (Optional) Datasphere Project limits configuration. The structure is documented below.
* `settings` - (Optional) Datasphere Project settings configuration. The structure is documented below.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Datasphere Project unique identifier
* `created_at` - Creation timestamp of the Yandex Datasphere Project.
* `created_by` - Creator account ID of the Yandex Datasphere Project.

---
The `limits` block supports:

* `max_units_per_hour` - (Optional) The number of units that can be spent per hour.
* `max_units_per_execution` - (Optional) The number of units that can be spent on the one execution.
* `balance` - (Optional) The number of units available to the project.
---

The `settings` block supports:

* `service_account_id` - (Optional) ID of the service account, on whose behalf all operations with clusters will be performed.
* `subnet_id` - (Optional) ID of the subnet where the DataProc cluster resides. Currently only subnets created in the availability zone ru-central1-a are supported.
* `data_proc_cluster_id` - (Optional) ID of the DataProc cluster.
* `commit_mode` - (Optional) Commit mode that is assigned to the project.
  * `STANDARD`: Commit happens after the execution of a cell or group of cells or after completion with an error. 
  * `AUTO`: Commit happens periodically. Also, automatic saving of state occurs when switching to another type of computing resource.
* `security_group_ids` - (Optional) List of network interfaces security groups.
* `ide` - (Optional) Project IDE. 
  * `JUPYTER_LAB`: Project running on JupyterLab IDE.
* `default_folder_id` - (Optional) Default project folder ID.
* `stale_exec_timeout_mode` - (Optional) Timeout to automatically stop stale executions.
  * `ONE_HOUR`: Setting to automatically stop stale execution after one hour with low consumption.
  * `THREE_HOURS`: Setting to automatically stop stale execution after three hours with low consumption.
  * `NO_TIMEOUT`: Setting to never automatically stop stale executions.

    
## Timeouts

This resource provides the following configuration options for timeouts:

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

A Datasphere Project can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_datasphere_project.default project_id
```
