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
resource "yandex_message_queue" "example_queue" {
  name                        = "ymq_terraform_example"
  visibility_timeout_seconds  = 600
  receive_wait_time_seconds   = 20
  message_retention_seconds   = 1209600
  redrive_policy              = jsonencode({
    deadLetterTargetArn = yandex_message_queue.example_deadletter_queue.arn
    maxReceiveCount     = 3
  })
}

resource "yandex_message_queue" "example_deadletter_queue" {
  name                        = "ymq_terraform_deadletter_example"
}
```

## FIFO queue

```hcl
resource "yandex_message_queue" "example_fifo_queue" {
  name                        = "ymq_terraform_fifo_example.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, forces new resource) Queue name. The maximum length is 80 characters. You can use numbers, letters, underscores, and hyphens in the name. The name of a FIFO queue must end with the `.fifo` suffix. If not specified, random name will be generated. Conflicts with `name_prefix`. For more information see [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/CreateQueue).

* `name_prefix` - (Optional, forces new resource) Generates random name with the specified prefix. Conflicts with `name`.

* `visibility_timeout_seconds` - (Optional) [Visibility timeout](https://cloud.yandex.com/docs/message-queue/concepts/visibility-timeout) for messages in a queue, specified in seconds. Valid values: from 0 to 43200 seconds (12 hours). Default: 30.

* `message_retention_seconds` - (Optional) The length of time in seconds to retain a message. Valid values: from 60 seconds (1 minute) to 1209600 seconds (14 days). Default: 345600 (4 days). For more information see [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/CreateQueue).

* `max_message_size` - (Optional) Maximum message size in bytes. Valid values: from 1024 bytes (1 KB) to 262144 bytes (256 KB). Default: 262144 (256 KB). For more information see [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/CreateQueue).

* `delay_seconds` - (Optional) Number of seconds to [delay the message from being available for processing](https://cloud.yandex.com/docs/message-queue/concepts/delay-queues#delay-queues). Valid values: from 0 to 900 seconds (15 minutes). Default: 0.

* `receive_wait_time_seconds` - (Optional) Wait time for the [ReceiveMessage](https://cloud.yandex.com/docs/message-queue/api-ref/message/ReceiveMessage) method (for long polling), in seconds. Valid values: from 0 to 20 seconds. Default: 0. For more information about long polling see [documentation](https://cloud.yandex.com/docs/message-queue/concepts/long-polling).

* `redrive_policy` - (Optional) Message redrive policy in [Dead Letter Queue](https://cloud.yandex.com/docs/message-queue/concepts/dlq). The source queue and DLQ must be the same type: for FIFO queues, the DLQ must also be a FIFO queue. For more information about redrive policy see [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/CreateQueue). Also you can use example in this page.

* `fifo_queue` - (Optional, forces new resource) Is this queue [FIFO](https://cloud.yandex.com/docs/message-queue/concepts/queue#fifo-queues). If this parameter is not used, a standard queue is created. You cannot change the parameter value for a created queue.

* `content_based_deduplication` - (Optional) Enables [content-based deduplication](https://cloud.yandex.com/docs/message-queue/concepts/deduplication#content-based-deduplication). Can be used only if queue is [FIFO](https://cloud.yandex.com/docs/message-queue/concepts/queue#fifo-queues).

* `access_key` - (Optional) The [access key](https://cloud.yandex.com/docs/iam/operations/sa/create-access-key) to use when applying changes. If omitted, `ymq_access_key` specified in provider config is used. For more information see [documentation](https://cloud.yandex.com/docs/message-queue/quickstart).

* `secret_key` - (Optional) The [secret key](https://cloud.yandex.com/docs/iam/operations/sa/create-access-key) to use when applying changes. If omitted, `ymq_secret_key` specified in provider config is used. For more information see [documentation](https://cloud.yandex.com/docs/message-queue/quickstart).

* `region_id` - (Optional, forces new resource) ID of the region where the message queue is located at.
The default is 'ru-central1'.

## Attributes Reference

Message Queue also has the following attributes:

* `id` - URL of the Yandex Message Queue.
* `arn` - ARN of the Yandex Message Queue. It is used for setting up a [redrive policy](https://cloud.yandex.com/docs/message-queue/concepts/dlq). See [documentation](https://cloud.yandex.com/docs/message-queue/api-ref/queue/SetQueueAttributes).

## Import

Yandex Message Queues can be imported using its `queue url`, e.g.

```
$ terraform import yandex_message_queue.example_import_queue https://message-queue.api.cloud.yandex.net/abcdefghijklmn123456/opqrstuvwxyz87654321/ymq_terraform_import_example
```
