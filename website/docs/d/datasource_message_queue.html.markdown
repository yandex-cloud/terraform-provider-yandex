---
layout: "yandex"
page_title: "Yandex: yandex_message_queue"
sidebar_current: "docs-yandex-datasource-message-queue"
description: |-
  Get information about a Yandex Message Queue.
---

# yandex\_message\_queue

Get information about a Yandex Message Queue. For more information about Yandex Message Queue, see
[Yandex.Cloud Message Queue](https://cloud.yandex.com/docs/message-queue).

## Example Usage

```hcl
data "yandex_message_queue" "example_queue" {
  name = "ymq_terraform_example"
}
```

## Argument Reference

* `name` - (Required) Queue name.
* `region_id` - (Optional) The region ID where the message queue is located.

## Attributes Reference

* `arn` - ARN of the queue. It is used for setting up a [redrive policy](https://cloud.yandex.com/docs/message-queue/concepts/dlq). See [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/SetQueueAttributes).
* `url` - URL of the queue.
