---
subcategory: "Cloud Functions"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Function Scaling Policy.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud Function Scaling Policies](https://cloud.yandex.com/docs/functions/)

## Example usage

{{ tffile "examples/function_scaling_policy/r_function_scaling_policy_1.tf" }}

## Argument Reference

The following arguments are supported:

* `function_id` (Required) - Yandex Cloud Function id used to define function

* `policy` - list definition for Yandex Cloud Function scaling policies
* `policy.#` - number of Yandex Cloud Function scaling policies
* `policy.{num}.tag` - Yandex Cloud Function version tag for Yandex Cloud Function scaling policy
* `policy.{num}.zone_instances_limit` - max number of instances in one zone for Yandex Cloud Function with tag
* `policy.{num}.zone_requests_limit` - max number of requests in one zone for Yandex Cloud Function with tag
