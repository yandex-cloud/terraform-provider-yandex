---
subcategory: "Datasphere"
page_title: "Yandex: yandex_datasphere_project"
description: |-
  Allows management of a Yandex Cloud Datasphere Project.
---

# yandex_datasphere_project (Resource)

Allows management of Yandex Cloud Datasphere Projects.

## Example usage

```terraform
//
// Create a new Datasphere Project.
//
resource "yandex_datasphere_project" "my-project" {
  name        = "example-datasphere-project"
  description = "Datasphere Project description"

  labels = {
    "foo" : "bar"
  }

  community_id = yandex_datasphere_community.my-community.id

  limits = {
    max_units_per_hour      = 10
    max_units_per_execution = 10
    balance                 = 10
  }

  settings = {
    service_account_id      = yandex_iam_service_account.my-account.id
    subnet_id               = yandex_vpc_subnet.my-subnet.id
    commit_mode             = "AUTO"
    data_proc_cluster_id    = "foo-data-proc-cluster-id"
    security_group_ids      = [yandex_vpc_security_group.my-security-group.id]
    ide                     = "JUPYTER_LAB"
    default_folder_id       = "foo-folder-id"
    stale_exec_timeout_mode = "ONE_HOUR"
  }
}

resource "yandex_datasphere_community" "my-community" {
  name               = "example-datasphere-community"
  description        = "Description of community"
  billing_account_id = "example-organization-id"
  labels = {
    "foo" : "bar"
  }
  organization_id = "example-organization-id"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `community_id` (String) Community ID where project would be created.
- `name` (String) The resource name.

### Optional

- `description` (String) The resource name.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `limits` (Attributes) Datasphere Project limits configuration. (see [below for nested schema](#nestedatt--limits))
- `settings` (Attributes) Datasphere Project settings configuration. (see [below for nested schema](#nestedatt--settings))
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `created_by` (String) Creator account ID of the Datasphere Project.
- `id` (String) The resource identifier.

<a id="nestedatt--limits"></a>
### Nested Schema for `limits`

Optional:

- `balance` (Number) The number of units available to the project.
- `max_units_per_execution` (Number) The number of units that can be spent on the one execution.
- `max_units_per_hour` (Number) The number of units that can be spent per hour.


<a id="nestedatt--settings"></a>
### Nested Schema for `settings`

Optional:

- `data_proc_cluster_id` (String) ID of the DataProcessing cluster.
- `default_folder_id` (String) Default project folder ID.
- `security_group_ids` (Set of String) The list of security groups applied to resource or their components.
- `service_account_id` (String) [Service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) which linked to the resource.
- `stale_exec_timeout_mode` (String) The timeout to automatically stop stale executions. The following modes can be used:
 * `ONE_HOUR`: Setting to automatically stop stale execution after one hour with low consumption.
  * `THREE_HOURS`: Setting to automatically stop stale execution after three hours with low consumption.
  * `NO_TIMEOUT`: Setting to never automatically stop stale executions.
- `subnet_id` (String) ID of the subnet where the DataProcessing cluster resides. Currently only subnets created in the availability zone `ru-central1-a` are supported.


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_datasphere_project.<resource Name> <resource Id>
terraform import yandex_datasphere_project.my-project ...
```
