---
subcategory: "Datasphere"
page_title: "Yandex: yandex_datasphere_community"
description: |-
  Allows management of a Yandex Cloud Datasphere Community.
---

# yandex_datasphere_community (Resource)

Allows management of Yandex Cloud Datasphere Communities.

## Example usage

```terraform
//
// Create a new Datasphere Community.
//
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

- `name` (String) The resource name.

### Optional

- `billing_account_id` (String) Billing account ID to associated with community
- `description` (String)
- `labels` (Map of String)
- `organization_id` (String) Organization ID where community would be created
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `created_by` (String) Creator account ID of the Datasphere Community
- `id` (String) The resource identifier.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_datasphere_community.<resource Name> <resource Id>
terraform import yandex_datasphere_community.my-community ...
```
