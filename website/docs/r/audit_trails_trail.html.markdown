---
layout: "yandex"
page_title: "Yandex: yandex_audit_trails_trail"
sidebar_current: "docs-yandex-audit-trails-trail"
description: |-
  Manages a trail resource.
---

# yandex\_audit\_trails\_trail

Allows management of [trail](https://cloud.yandex.ru/en/docs/audit-trails/concepts/trail)

## Example usage

```hcl
resource "yandex_audit_trails_trail" "basic_trail" {
  name = "a-trail"
  folder_id = "home-folder"
  description = "Some trail description"
  
  labels = {
    key = "value"
  }
  
  service_account_id = "trail-service-account"
  
  logging_destination {
    log_group_id = "some-log-group"
  }
  
  filter {
    path_filter {
      any_filter {
        resource_id = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    event_filters {
      service = "storage"
      categories {
        plane: "DATA_PLANE"
        type: "WRITE"
      }
      path_filter {
        any_filter {
          resource_id = "home-folder"
          resource_type = "resource-manager.folder"
        }
      }
    }
    event_filters {
      service = "dns"
      categories {
        plane: "DATA_PLANE"
        type: "READ"
      }
      path_filter {
        some_filter {
          resource_id = "home-folder"
          resource_type = "resource-manager.folder"
          any_filters {
            resource_id = "vpc-net-id-1"
            resource_type = "vpc.network"
          }
          any_filters {
            resource_id = "vpc-net-id-2"
            resource_type = "vpc.network"
          }
        }
      }
    }
  }
}
```

## Argument Reference

* `name` - (Required) Name of the trail.

* `folder_id` - (Required) ID of the folder to which the trail belongs.

* `description` - (Optional) Description of the trail.

* `labels` - (Optional) Labels defined by the user.

* `service_account_id` - (Required) ID of the [IAM service account](https://cloud.yandex.ru/en/docs/iam/concepts/users/service-accounts) that is used by the trail.

* `storage_destination` - Structure describing destination bucket of the trail. Mutually exclusive with `logging_destination` and `data_stream_destination`.

  * `bucket_name` - (Required) Name of the [destination bucket](https://cloud.yandex.ru/en/docs/storage/concepts/bucket)

  * `object_prefix` - (Optional) Additional prefix of the uploaded objects. If not specified, objects will be uploaded with prefix equal to `trail_id`

* `logging_destination` - Structure describing destination log group of the trail. Mutually exclusive with `storage_destination` and `data_stream_destination`.

  * `log_group_id` - (Required) ID of the destination [Cloud Logging Group](https://cloud.yandex.ru/ru/docs/logging/concepts/log-group)

* `data_stream_destination` - Structure describing destination data stream of the trail. Mutually exclusive with `logging_destination` and `storage_destination`.

  * `database_id` - (Required) ID of the [YDB](https://cloud.yandex.ru/ru/docs/ydb/concepts/resources) hosting the destination data stream.

  * `stream_name` - (Required) Name of the [YDS stream](https://cloud.yandex.ru/ru/docs/data-streams/concepts/glossary#stream-concepts) belonging to the specified YDB.

* `filter` - Structure describing event filtering process for the trail.

  * `path_filter` - (Optional) Structure describing filtering process for default control plane events. If omitted, the trail will not deliver this category

    * `any_filter` - Structure describing that events will be gathered from all cloud resources that belong to the parent resource. Mutually exclusive with `some_filter`.

      * `resource_id` - (Required) ID of the parent resource.

      * `resource_type` - (Required) Resource type of the parent resource.

    * `some_filter` - Structure describing that events will be gathered from some of the cloud resources that belong to the parent resource. Mutually exclusive with `any_filter`.

      * `resource_id` - (Required) ID of the parent resource.

      * `resource_type` - (Required) Resource type of the parent resource.

      * `any_filters` - (Required) List of child resources from which events will be gathered

        * `resource_id` - (Required) ID of the child resource.

        * `resource_type` - (Required) Resource type of the child resource.

  * `event_filters` - Structure describing filtering process for the service-specific data plane events

    * `service` - (Required) ID of the service which events will be gathered

    * `categories` - (Required) List of structures describing categories of gathered data plane events

      * `plane` - (Required) Type of the event by its relation to the cloud resource model. Possible values: `CONTROL_PLANE`/`DATA_PLANE`

      * `type` - (Required) Type of the event by its operation effect on the resource. Possible values: `READ`/`WRITE`

    * `path_filter` - (Required) Structure describing filtering process based on cloud resources for the described event set. Structurally equal to the `filter.path_filter`

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - Status of this trail.
* `trail_id` - ID of the trail resource.

## Timeouts

`yandex_audit_trails_trail` provides the following configuration options for
[timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 5 minutes.
- `update` - Default 5 minutes.
- `delete` - Default 5 minutes.

## Import

A trail can be imported using the `id` of the resource, e.g.

```bash
$ terraform import yandex_audit_trails_trail.infosec-trail trail_id
```