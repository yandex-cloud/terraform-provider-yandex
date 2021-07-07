---
layout: "yandex"
page_title: "Yandex: yandex_function_scaling_policy"
sidebar_current: "docs-yandex-datasource-yandex-function-scaling-policy"
description: |-
  Get information about a Yandex Cloud Functions Scaling Policy.
---

# yandex\_function\_scaling\_policy

Get information about a Yandex Cloud Function Scaling Policy. For more information about Yandex Cloud Functions, see 
[Yandex Cloud Functions](https://cloud.yandex.com/docs/functions/).

```hcl
data "yandex_function_scaling_policy" "my_scaling_policy" {
  function_id = "are1samplefunction11"
}
```

This data source is used to define [Yandex Cloud Function Scaling Policy](https://cloud.yandex.com/docs/functions/) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `function_id` (Required) - Yandex Cloud Function id used to define function

## Attributes Reference

The following attributes are exported:

* `policy` - list definition for Yandex Cloud Function scaling policies
* `policy.#` - number of Yandex Cloud Function scaling policies
* `policy.{num}.tag` - Yandex.Cloud Function version tag for Yandex Cloud Function scaling policy
* `policy.{num}.zone_instances_limit` - max number of instances in one zone for Yandex.Cloud Function with tag
* `policy.{num}.zone_requests_limit` - max number of requests in one zone for Yandex.Cloud Function with tag