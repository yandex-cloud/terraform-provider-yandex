---
subcategory: "Message Queue"
page_title: "Yandex: yandex_message_queue"
description: |-
  Get information about a Yandex Message Queue.
---

# yandex_message_queue (Data Source)

Get information about a Yandex Message Queue. For more information about Yandex Message Queue, see [Yandex Cloud Message Queue](https://yandex.cloud/docs/message-queue).

## Example usage

```terraform
//
// Get information about existing Message Queue.
//
data "yandex_message_queue" "my_queue" {
  name = "ymq_terraform_example"
}
```

## Argument Reference

* `name` - (Required) Queue name.
* `region_id` - (Optional) The region ID where the message queue is located.

## Attributes Reference

* `arn` - ARN of the queue. It is used for setting up a [redrive policy](https://yandex.cloud/docs/message-queue/concepts/dlq). See [documentation](https://yandex.cloud/docs/message-queue/api-ref/queue/SetQueueAttributes).
* `url` - URL of the queue.
