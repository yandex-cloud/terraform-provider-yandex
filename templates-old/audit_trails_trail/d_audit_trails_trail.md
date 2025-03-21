---
subcategory: "Audit Trails"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a trail.
---

# {{.Name}} ({{.Type}})

Get information about a trail. For information about the trail concept, see [official documentation](https://yandex.cloud/docs/audit-trails/concepts/trail)

## Example usage

{{ tffile "examples/audit_trails_trail/d_audit_trails_trail_1.tf" }}

## Argument Reference

The following arguments are supported:

* `trail_id` - (Required) trail ID.

## Attributes Reference

The following attributes are exported:

* `name` - Name of the trail.

* `folder_id` - ID of the folder to which the trail belongs.

* `description` - (Optional) Description of the trail.

* `labels` - (Optional) Labels defined by the user.

* `service_account_id` - ID of the [IAM service account](https://yandex.cloud/docs/iam/concepts/users/service-accounts) that is used by the trail.

* `status` - Status of the trail.

* `storage_destination` - Structure describing destination bucket of the trail. Mutually exclusive with `logging_destination` and `data_stream_destination`.

  * `bucket_name` - Name of the [destination bucket](https://yandex.cloud/docs/storage/concepts/bucket).

  * `object_prefix` - (Optional) Additional prefix of the uploaded objects. If not specified, objects will be uploaded with prefix equal to `trail_id`.

* `logging_destination` - Structure describing destination log group of the trail. Mutually exclusive with `storage_destination` and `data_stream_destination`.

  * `log_group_id` - ID of the destination [Cloud Logging Group](https://yandex.cloud/docs/logging/concepts/log-group).

* `data_stream_destination` - Structure describing destination data stream of the trail. Mutually exclusive with `logging_destination` and `storage_destination`.

  * `database_id` - ID of the [YDB](https://yandex.cloud/docs/ydb/concepts/resources) hosting the destination data stream.

  * `stream_name` - Name of the [YDS stream](https://yandex.cloud/docs/data-streams/concepts/glossary#stream-concepts) belonging to the specified YDB.

* `filter` - (Optional) Structure describing event filtering process for the trail. Deprecated. Use `filtering_policy` instead. Mutually exclusive with `filtering_policy`.

  * `path_filter` - (Optional) Structure describing filtering process for default control plane events. If omitted, the trail will not deliver this category.

    * `any_filter` - Structure describing that events will be gathered from all cloud resources that belong to the parent resource. Mutually exclusive with `some_filter`.

      * `resource_id` - ID of the parent resource.

      * `resource_type` - Resource type of the parent resource.

    * `some_filter` - Structure describing that events will be gathered from some of the cloud resources that belong to the parent resource. Mutually exclusive with `any_filter`.

      * `resource_id` - ID of the parent resource.

      * `resource_type` - Resource type of the parent resource.

      * `any_filters` - List of child resources from which events will be gathered.

        * `resource_id` - ID of the child resource.

        * `resource_type` - Resource type of the child resource.

  * `event_filters` - Structure describing filtering process for the service-specific data plane events.

    * `service` - ID of the service which events will be gathered.

    * `categories` - List of structures describing categories of gathered data plane events.

      * `plane` - Type of the event by its relation to the cloud resource model. Possible values: `CONTROL_PLANE`/`DATA_PLANE`.

      * `type` - Type of the event by its operation effect on the resource. Possible values: `READ`/`WRITE`.

    * `path_filter` - Structure describing filtering process based on cloud resources for the described event set. Structurally equal to the `filter.path_filter`.

* `filtering_policy` - (Optional) Structure describing event filtering process for the trail. Mutually exclusive with `filter`. At least one of the `management_events_filter` or `data_events_filter` fields will be filled.

  * `management_events_filter` - (Optional) Structure describing filtering process for management events.

    * `resource_scope` - Structure describing that events will be gathered from the specified resource.

      * `resource_id` - ID of the monitored resource.

      * `resource_type` - Resource type of the monitored resource.

  * `data_events_filter` - (Optional) Structure describing filtering process for the service-specific data events.

    * `service` - ID of the service which events will be gathered.

    * `resource_scope` - Structure describing that events will be gathered from the specified resource.

      * `resource_id` - ID of the monitored resource.

      * `resource_type` - Resource type of the monitored resource.

    * `included_events` - (Optional) A list of events that will be gathered by the trail from this service. New events won't be gathered by default when this option is specified. Mutually exclusive with `excluded_events`

    * `excluded_events` - (Optional) A list of events that won't be gathered by the trail from this service. New events will be automatically gathered when this option is specified. Mutually exclusive with `included_events`
