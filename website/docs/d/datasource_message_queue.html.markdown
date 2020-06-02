---
layout: "yandex"
page_title: "Yandex: yandex_message_queue"
sidebar_current: "docs-yandex-datasource-message-queue"
description: |-
  Get information about a Yandex Message Queue.
---

# yandex\_message\_queue

Use this data source to get the ARN and URL of Yandex Message Queue.
By using this data source, you can reference message queues without having to hardcode
the ARNs as input.

## Example Usage

```hcl
data "yandex_message_queue" "example" {
  name = "queue"
}
```

## Argument Reference

* `name` - (Required) The name of the queue to match.

## Attributes Reference

* `arn` - ARN of the queue. It is used for setting up a [redrive policy](https://cloud.yandex.com/docs/message-queue/concepts/dlq). See [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/SetQueueAttributes).
* `url` - The URL of the queue.
