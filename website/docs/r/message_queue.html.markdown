---
layout: "yandex"
page_title: "Yandex: yandex_message_queue"
sidebar_current: "docs-yandex-message-queue"
description: |-
  Allows management of a Yandex.Cloud Message Queue.
---

# yandex\_message\_queue

Allows management of [Yandex.Cloud Message Queue](https://cloud.yandex.com/docs/message-queue).

## Example Usage

```hcl
resource "yandex_message_queue" "terraform_queue" {
  name                      = "terraform-example-queue"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 20
  redrive_policy            = jsonencode({
    deadLetterTargetArn = yandex_message_queue.terraform_queue_deadletter.arn
    maxReceiveCount     = 4
  })
}

resource "yandex_message_queue" "terraform_queue_deadletter" {
  name                      = "terraform-example-dlq-queue"
}
```

## FIFO queue

```hcl
resource "yandex_message_queue" "terraform_queue" {
  name                        = "terraform-example-queue.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) This is the human-readable name of the queue. If omitted, Terraform will assign a random name.

* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.

* `visibility_timeout_seconds` - (Optional) The visibility timeout for the queue. An integer from 0 to 43200 (12 hours). The default for this attribute is 30. For more information about visibility timeout, see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/visibility-timeout).

* `message_retention_seconds` - (Optional) The number of seconds Message Queue retains a message. Integer representing seconds, from 60 (1 minute) to 1209600 (14 days). The default for this attribute is 345600 (4 days).

* `max_message_size` - (Optional) The limit of how many bytes a message can contain. An integer from 1024 bytes (1 KiB) up to 262144 bytes (256 KiB). The default for this attribute is 262144 (256 KiB).

* `delay_seconds` - (Optional) The time in seconds that the delivery of all messages in the queue will be delayed. An integer from 0 to 900 (15 minutes). The default for this attribute is 0 seconds. For more information about delay, see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/delay-queues).

* `receive_wait_time_seconds` - (Optional) The time for which a ReceiveMessage call will wait for a message to arrive (long polling) before returning. An integer from 0 to 20 (seconds). The default for this attribute is 0, meaning that the call will return immediately. For more information about long polling see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/long-polling).

* `redrive_policy` - (Optional) The JSON policy to set up the Dead Letter Queue. For more information see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/dlq). Also you can use example in this page.

* `fifo_queue` - (Optional, Forces new resource) Boolean designating a FIFO queue. If not set, it defaults to `false` making it standard.

* `content_based_deduplication` - (Optional) Enables content-based deduplication for FIFO queues. For more information, see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/deduplication).

* `access_key` - (Optional) The access key to use when applying changes. If omitted, `ymq_access_key` specified in provider config is used.

* `secret_key` - (Optional) The secret key to use when applying changes. If omitted, `ymq_secret_key` specified in provider config is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The URL for the created Yandex Message queue.
* `arn` - The ARN of the message queue. It is used for setting up a [redrive policy](https://cloud.yandex.com/docs/message-queue/concepts/dlq). See [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/SetQueueAttributes).

## Import

Yandex Message Queues can be imported using the `queue url`, e.g.

```
$ terraform import yandex_message_queue.public_queue https://message-queue.api.cloud.yandex.net/abcdefghijklmn123456/opqrstuvwxyz87654321/example-queue
```
