---
subcategory: "Serverless Cloud Functions"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Functions Trigger.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud Functions Trigger](https://yandex.cloud/docs/functions/)

## Example usage

{{ tffile "examples/function_trigger/r_function_trigger_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` (Required) - Yandex Cloud Functions Trigger name used to define trigger
* `folder_id` - (Optional) Folder ID for the Yandex Cloud Functions Trigger
* `description` - Description of the Yandex Cloud Functions Trigger
* `labels` - A set of key/value label pairs to assign to the Yandex Cloud Functions Trigger

* `function` - [Yandex Cloud Function](https://yandex.cloud/docs/functions/concepts/function) settings definition for Yandex Cloud Functions Trigger
* `function.0.id` - Yandex Cloud Function ID for Yandex Cloud Functions Trigger
* `function.0.service_account_id` - Service account ID for Yandex Cloud Function for Yandex Cloud Functions Trigger
* `function.0.tag` - Tag for Yandex Cloud Function for Yandex Cloud Functions Trigger
* `function.0.retry_attempts` - Retry attempts for Yandex Cloud Function for Yandex Cloud Functions Trigger
* `function.0.retry_interval` - Retry interval in seconds for Yandex Cloud Function for Yandex Cloud Functions Trigger

* `container` - [Yandex Cloud Serverless Container](https://yandex.cloud/docs/serverless-containers/concepts/container) settings definition for Yandex Cloud Functions Trigger
* `container.0.id` - Yandex Cloud Serverless Container ID for Yandex Cloud Functions Trigger
* `container.0.service_account_id` - Service account ID for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.path` - Path for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.retry_attempts` - Retry attempts for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger
* `container.0.retry_interval` - Retry interval in seconds for Yandex Cloud Serverless Container for Yandex Cloud Functions Trigger

* `dlq` - Dead Letter Queue settings definition for Yandex Cloud Functions Trigger
* `dlq.0.queue_id` - ID of Dead Letter Queue for Trigger (Queue ARN)
* `dlq.0.service_account_id` - Service Account ID for Dead Letter Queue for Yandex Cloud Functions Trigger

* `iot` - [IoT](https://yandex.cloud/docs/functions/concepts/trigger/iot-core-trigger) settings definition for Yandex Cloud Functions Trigger, if present. Only one section `iot` or `message_queue` or `object_storage` or `timer` can be defined.
* `iot.0.registry_id` - IoT Registry ID for Yandex Cloud Functions Trigger
* `iot.0.device_id` - IoT Device ID for Yandex Cloud Functions Trigger
* `iot.0.topic` - IoT Topic for Yandex Cloud Functions Trigger
* `iot.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `iot.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

* `message_queue` - [Message Queue](https://yandex.cloud/docs/functions/concepts/trigger/ymq-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `message_queue.0.queue_id` - Message Queue ID for Yandex Cloud Functions Trigger
* `message_queue.0.service_account_id` - Message Queue Service Account ID for Yandex Cloud Functions Trigger
* `message_queue.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `message_queue.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger
* `message_queue.0.visibility_timeout` - Visibility timeout for Yandex Cloud Functions Trigger

* `object_storage` - [Object Storage](https://yandex.cloud/docs/functions/concepts/trigger/os-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `object_storage.0.bucket_id` - Object Storage Bucket ID for Yandex Cloud Functions Trigger
* `object_storage.0.prefix` - Prefix for Object Storage for Yandex Cloud Functions Trigger
* `object_storage.0.suffix` - Suffix for Object Storage for Yandex Cloud Functions Trigger
* `object_storage.0.create` - Boolean flag for setting create event for Yandex Cloud Functions Trigger
* `object_storage.0.update` - Boolean flag for setting update event for Yandex Cloud Functions Trigger
* `object_storage.0.delete` - Boolean flag for setting delete event for Yandex Cloud Functions Trigger
* `object_storage.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `object_storage.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

* `timer` - [Timer](https://yandex.cloud/docs/functions/concepts/trigger/timer) settings definition for Yandex Cloud Functions Trigger, if present
* `timer.0.cron_expression` - Cron expression for timer for Yandex Cloud Functions Trigger
* `timer.0.payload` - Payload to be passed to function

* `logging` - [Logging](https://yandex.cloud/docs/functions/concepts/trigger/cloud-logging-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `logging.0.group_id` - Logging group ID for Yandex Cloud Functions Trigger
* `logging.0.resource_ids` - Resource ID filter setting for Yandex Cloud Functions Trigger
* `logging.0.resource_types` - Resource type filter setting for Yandex Cloud Functions Trigger
* `logging.0.levels` - Logging level filter setting for Yandex Cloud Functions Trigger
* `logging.0.stream_names` - Logging stream name filter setting for Yandex Cloud Functions Trigger
* `logging.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `logging.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

* `container_registry` - [Container Registry](https://yandex.cloud/docs/functions/concepts/trigger/cr-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `container_registry.0.registry_id` - Container Registry ID for Yandex Cloud Functions Trigger
* `container_registry.0.image_name` - Image name filter setting for Yandex Cloud Functions Trigger
* `container_registry.0.tag` - Image tag filter setting for Yandex Cloud Functions Trigger
* `container_registry.0.create_image` - Boolean flag for setting create image event for Yandex Cloud Functions Trigger
* `container_registry.0.delete_image` - Boolean flag for setting delete image event for Yandex Cloud Functions Trigger
* `container_registry.0.create_image_tag` - Boolean flag for setting create image tag event for Yandex Cloud Functions Trigger
* `container_registry.0.delete_image_tag` - Boolean flag for setting delete image tag event for Yandex Cloud Functions Trigger
* `container_registry.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `container_registry.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

* `data_streams` - [Data Streams](https://yandex.cloud/docs/functions/concepts/trigger/data-streams-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `data_streams.0.stream_name` - Stream name for Yandex Cloud Functions Trigger
* `data_streams.0.database` - Stream database for Yandex Cloud Functions Trigger
* `data_streams.0.service_account_id` - Service account ID to access data stream for Yandex Cloud Functions Trigger
* `data_streams.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `data_streams.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

* `mail` - [Mail](https://yandex.cloud/docs/functions/concepts/trigger/mail-trigger) settings definition for Yandex Cloud Functions Trigger, if present
* `mail.0.attachments_bucket_id` - Object Storage Bucket ID for Yandex Cloud Functions Trigger
* `mail.0.service_account_id` - Service account ID to access object storage for Yandex Cloud Functions Trigger
* `mail.0.batch_cutoff` - Batch Duration in seconds for Yandex Cloud Functions Trigger
* `mail.0.batch_size` - Batch Size for Yandex Cloud Functions Trigger

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the Yandex Cloud Functions Trigger

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "shell" "examples/function_trigger/import.sh" }}
