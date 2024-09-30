---
subcategory: "Serverless Function"
page_title: "Yandex: yandex_function_scaling_policy"
description: |-
  Allows management of a Yandex Cloud Function Scaling Policy.
---


# yandex_function_scaling_policy




Allows management of [Yandex Cloud Function Scaling Policies](https://cloud.yandex.com/docs/functions/)

```terraform
resource "yandex_function_trigger" "my_trigger" {
  name        = "some_name"
  description = "any description"
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id = "tf-test"
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
