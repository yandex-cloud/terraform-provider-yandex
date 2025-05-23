---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: yandex_function_scaling_policy"
description: |-
  Get information about a Yandex Cloud Functions Scaling Policy.
---

# yandex_function_scaling_policy (Data Source)

Get information about a Yandex Cloud Function Scaling Policy. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://yandex.cloud/docs/functions/).

This data source is used to define [Yandex Cloud Function Scaling Policy](https://yandex.cloud/docs/functions/) that can be used by other resources.

## Example usage

```terraform
//
// Get information about existing Cloud Function Scaling Policy.
//
data "yandex_function_scaling_policy" "my_scaling_policy" {
  function_id = "d4e45**********pqvd3"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `function_id` (String) Yandex Cloud Function id used to define function.

### Optional

- `policy` (Block List) (see [below for nested schema](#nestedblock--policy))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--policy"></a>
### Nested Schema for `policy`

Optional:

- `zone_instances_limit` (Number) Max number of instances in one zone for Yandex Cloud Function with tag.

- `zone_requests_limit` (Number) Max number of requests in one zone for Yandex Cloud Function with tag.


Read-Only:

- `tag` (String) Yandex Cloud Function version tag for Yandex Cloud Function scaling policy.

