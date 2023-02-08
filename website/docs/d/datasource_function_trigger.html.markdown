---
layout: "yandex"
page_title: "Yandex: yandex_function_trigger"
sidebar_current: "docs-yandex-datasource-yandex-function-trigger"
description: |-
  Get information about a Yandex Cloud Functions Trigger.
---

# yandex\_function\_trigger

Get information about a Yandex Cloud Function Trigger. For more information about Yandex Cloud Functions, see 
[Yandex Cloud Functions](https://cloud.yandex.com/docs/functions/).

```hcl
data "yandex_function_trigger" "my_trigger" {
  trigger_id = "are1sampletrigger11"
}
```

This data source is used to define [Yandex Cloud Functions Trigger](https://cloud.yandex.com/docs/functions/concepts/trigger) that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `trigger_id` (Optional) - Yandex Cloud Functions Trigger id used to define trigger

* `name` (Optional) - Yandex Cloud Functions Trigger name used to define trigger

* `folder_id` - (Optional) Folder ID for the Yandex Cloud Functions Trigger

~> **NOTE:** Either `trigger_id` or `name` must be specified.

## Attributes Reference

The following attributes are exported:

* `description` - Description of the Yandex Cloud Functions Trigger
* `folder_id` - Folder ID for the Yandex Cloud Functions Trigger
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud Functions Trigger
* `created_at` - Creation timestamp of the Yandex Cloud Functions Trigger

* `function` - [Yandex.Cloud Function](https://cloud.yandex.com/docs/functions/concepts/function) settings definition for Yandex Cloud Functions Trigger
* `function.0.id` - Yandex.Cloud Function ID for Yandex Cloud Functions Trigger
* `function.0.service_account_id` - Service account ID for Yandex.Cloud Function for Yandex Cloud Functions Trigger
* `function.0.tag` - Tag for Yandex.Cloud Function for Yandex Cloud Functions Trigger
* `function.0.retry_attempts` - Retry attempts for Yandex.Cloud Function for Yandex Cloud Functions Trigger
* `function.0.retry_interval` - Retry interval in seconds for Yandex.Cloud Function for Yandex Cloud Functions Trigger

* `` - [Yandex.Cloud Serverless Container](https://cloud.yandex.com/en-ru/docs/serverless-containers/concepts/container) settings definition for Yandex Cloud Functions Trigger
* `container.0.id` - Yandex.Cloud Serverless Container ID for Yandex Cloud Functions Trigger
* `container.0.service_account_id` - Service account ID for Yandex.Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.path` - Path for Yandex.Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.retry_attempts` - Retry attempts for Yandex.Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.retry_interval` - Retry interval in seconds for Yandex.Cloud Serverless Container for Yandex Cloud Functions Trigger

* `dlq` - Dead Letter Queue settings definition for Yandex Cloud Functions Trigger
* `dlq.0.queue_id` - ID of Dead Letter Queue for Trigger (Queue ARN)
* `dlq.0.service_account_id` - Service Account ID for Dead Letter Queue for Yandex Cloud Functions Trigger

* `iot` - [IoT](https://cloud.yandex.com/docs/functions/concepts/trigger/iot-core-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `iot.0.registry_id` - IoT Registry ID for Yandex Cloud Functions Trigger
* `iot.0.device_id` - IoT Device ID for Yandex Cloud Functions Trigger
* `iot.0.topic` - IoT Topic for Yandex Cloud Functions Trigger

* `message_queue` - [Message Queue](https://cloud.yandex.com/docs/functions/concepts/trigger/ymq-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `message_queue.0.queue_id` - Message Queue ID for Yandex Cloud Functions Trigger
* `message_queue.0.service_account_id` - Message Queue Service Account ID for Yandex Cloud Functions Trigger
* `message_queue.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `message_queue.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger
* `message_queue.0.visibility_timeout` - Visibility timeout for Yandex Cloud Functions Trigger

* `object_storage` - [Object Storage](https://cloud.yandex.com/docs/functions/concepts/trigger/os-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `object_storage.0.bucket_id` - Object Storage Bucket ID for Yandex Cloud Functions Trigger
* `object_storage.0.prefix` - Prefix for Object Storage for Yandex Cloud Functions Trigger
* `object_storage.0.suffix` - Suffix for Object Storage for Yandex Cloud Functions Trigger
* `object_storage.0.create` - Boolean flag for setting create event for Yandex Cloud Functions Trigger
* `object_storage.0.update` - Boolean flag for setting update event for Yandex Cloud Functions Trigger
* `object_storage.0.delete` - Boolean flag for setting delete event for Yandex Cloud Functions Trigger

* `timer` - [Timer](https://cloud.yandex.com/docs/functions/concepts/trigger/timer) settings definition for Yandex Cloud Functions Trigger, if present
* `timer.0.cron_expression` - Cron expression for timer for Yandex Cloud Functions Trigger



