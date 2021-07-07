---
layout: "yandex"
page_title: "Yandex: yandex_function_scaling_policy"
sidebar_current: "docs-yandex-function-scaling-policy"
description: |-
 Allows management of a Yandex Cloud Function Scaling Policy.
---

# yandex\_function\_scaling\_policy

Allows management of [Yandex Cloud Function Scaling Policies](https://cloud.yandex.com/docs/functions/)

## Example Usage

```hcl
resource "yandex_function_scaling_policy" "my_scaling_policy" {
  function_id = "are1samplefunction11"
  policy {
    tag = "$latest"
    zone_instances_limit = 3
    zone_requests_limit  = 100
  }
  policy {
    tag = "my_tag"
    zone_instances_limit = 4
    zone_requests_limit  = 150
  }
}
```

## Argument Reference

The following arguments are supported:

* `function_id` (Required) - Yandex Cloud Function id used to define function

* `policy` - list definition for Yandex Cloud Function scaling policies
* `policy.#` - number of Yandex Cloud Function scaling policies
* `policy.{num}.tag` - Yandex.Cloud Function version tag for Yandex Cloud Function scaling policy
* `policy.{num}.zone_instances_limit` - max number of instances in one zone for Yandex.Cloud Function with tag
* `policy.{num}.zone_requests_limit` - max number of requests in one zone for Yandex.Cloud Function with tag