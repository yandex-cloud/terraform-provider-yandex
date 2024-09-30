---
subcategory: "Audit Trails"
page_title: "Yandex: yandex_audit_trails_trail"
description: |-
  Manages a trail resource.
---


# yandex_audit_trails_trail




Allows management of [trail](https://cloud.yandex.ru/en/docs/audit-trails/concepts/trail)

```terraform
resource "yandex_audit_trails_trail" "basic_trail" {
  name        = "a-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  logging_destination {
    log_group_id = "some-log-group"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "storage"
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "dns"
      resource_scope {
        resource_id   = "vpc-net-id-1"
        resource_type = "vpc.network"
      }
      resource_scope {
        resource_id   = "vpc-net-id-2"
        resource_type = "vpc.network"
      }
    }
  }
}
```

Trail delivering events to YDS and gathering such events:

* Management events from the 'some-organization' organization
* DNS data events from the 'some-organization' organization
* Object Storage data events from the 'some-organization' organization

```terraform
resource "yandex_audit_trails_trail" "basic_trail" {
  name        = "a-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  data_stream_destination {
    database_id = "some-database"
    stream_name = "some-stream"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
    data_events_filter {
      service = "storage"
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
    data_events_filter {
      service = "dns"
      resource_scope {
        resource_id   = "some-organization"
        resource_type = "organization-manager.organization"
      }
    }
  }
}
```

Trail delivering events to Object Storage and gathering such events:

* Management events from the 'home-folder' folder
* Managed PostgreSQL data events from the 'home-folder' folder

```terraform
resource "yandex_audit_trails_trail" "basic_trail" {
  name        = "a-trail"
  folder_id   = "home-folder"
  description = "Some trail description"

  labels = {
    key = "value"
  }

  service_account_id = "trail-service-account"

  storage_destination {
    bucket_name   = "some-bucket"
    object_prefix = "some-prefix"
  }

  filtering_policy {
    management_events_filter {
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
      }
    }
    data_events_filter {
      service = "mdb.postgresql"
      resource_scope {
        resource_id   = "home-folder"
        resource_type = "resource-manager.folder"
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

  * `bucket_name` - (Required) Name of the [destination bucket](https://cloud.yandex.ru/en/docs/storage/concepts/bucket).

  * `object_prefix` - (Optional) Additional prefix of the uploaded objects. If not specified, objects will be uploaded with prefix equal to `trail_id`.

* `logging_destination` - Structure describing destination log group of the trail. Mutually exclusive with `storage_destination` and `data_stream_destination`.

  * `log_group_id` - (Required) ID of the destination [Cloud Logging Group](https://cloud.yandex.ru/ru/docs/logging/concepts/log-group).

* `data_stream_destination` - Structure describing destination data stream of the trail. Mutually exclusive with `logging_destination` and `storage_destination`.

  * `database_id` - (Required) ID of the [YDB](https://cloud.yandex.ru/ru/docs/ydb/concepts/resources) hosting the destination data stream.

  * `stream_name` - (Required) Name of the [YDS stream](https://cloud.yandex.ru/ru/docs/data-streams/concepts/glossary#stream-concepts) belonging to the specified YDB.

* `filter` - Structure describing event filtering process for the trail.

  * `path_filter` - (Optional) Structure describing filtering process for default control plane events. If omitted, the trail will not deliver this category.

    * `any_filter` - Structure describing that events will be gathered from all cloud resources that belong to the parent resource. Mutually exclusive with `some_filter`.

      * `resource_id` - (Required) ID of the parent resource.

      * `resource_type` - (Required) Resource type of the parent resource.

    * `some_filter` - Structure describing that events will be gathered from some of the cloud resources that belong to the parent resource. Mutually exclusive with `any_filter`.

      * `resource_id` - (Required) ID of the parent resource.

      * `resource_type` - (Required) Resource type of the parent resource.

      * `any_filters` - (Required) List of child resources from which events will be gathered.

        * `resource_id` - (Required) ID of the child resource.

        * `resource_type` - (Required) Resource type of the child resource.

  * `event_filters` - Structure describing filtering process for the service-specific data plane events.

    * `service` - (Required) ID of the service which events will be gathered.

    * `categories` - (Required) List of structures describing categories of gathered data plane events.

      * `plane` - (Required) Type of the event by its relation to the cloud resource model. Possible values: `CONTROL_PLANE`/`DATA_PLANE`.

      * `type` - (Required) Type of the event by its operation effect on the resource. Possible values: `READ`/`WRITE`.

    * `path_filter` - (Required) Structure describing filtering process based on cloud resources for the described event set. Structurally equal to the `filter.path_filter`.

  * `filtering_policy` - (Optional) Structure describing event filtering process for the trail. Mutually exclusive with `filter`. At least one of the `management_events_filter` or `data_events_filter` fields will be filled.

    * `management_events_filter` - (Optional) Structure describing filtering process for management events.

      * `resource_scope` - (Required) Structure describing that events will be gathered from the specified resource.

        * `resource_id` - (Required) ID of the monitored resource.

        * `resource_type` - (Required) Resource type of the monitored resource.

    * `data_events_filter` - (Optional) Structure describing filtering process for the service-specific data events.

      * `service` - (Required) ID of the service which events will be gathered.

      * `resource_scope` - (Required) Structure describing that events will be gathered from the specified resource.

        * `resource_id` - (Required) ID of the monitored resource.

        * `resource_type` - (Required) Resource type of the monitored resource.

      * `included_events` - (Optional) A list of events that will be gathered by the trail from this service. New events won't be gathered by default when this option is specified. Mutually exclusive with `excluded_events`.

      * `excluded_events` - (Optional) A list of events that won't be gathered by the trail from this service. New events will be automatically gathered when this option is specified. Mutually exclusive with `included_events`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - Status of this trail.
* `trail_id` - ID of the trail resource.

## Timeouts

`yandex_audit_trails_trail` provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default 5 minutes.
- `update` - Default 5 minutes.
- `delete` - Default 5 minutes.

## Import

A trail can be imported using the `id` of the resource, e.g.

```bash
$ terraform import yandex_audit_trails_trail.infosec-trail trail_id
```

## Migration from deprecated filter field

In order to migrate from unsing `filter` to the `filtering_policy`, you will have to:

* Remove the `filter.event_filters.categories` blocks. With the introduction of `included_events`/`excluded_events` you can configure filtering per each event type.

* Replace the `filter.event_filters.path_filter` with the appropriate `resource_scope` blocks. You have to account that `resource_scope` does not support specifying relations between resources, so your configuration will simplify to only the actual resources, that will be monitored.

Before

```terraform
event_filters {
  path_filter {
    some_filter {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
      any_filters {
        resource_id   = "vpc-net-id-1"
        resource_type = "vpc.network"
      }
      any_filters {
        resource_id   = "vpc-net-id-2"
        resource_type = "vpc.network"
      }
    }
  }
}
```

After

```terraform
data_events_filter {
  service = "dns"
  resource_scope {
    resource_id   = "vpc-net-id-1"
    resource_type = "vpc.network"
  }
  resource_scope {
    resource_id   = "vpc-net-id-2"
    resource_type = "vpc.network"
  }
}
```

* Replace the `filter.path_filter` block with the `filtering_policy.management_events_filter`. New API states management events filtration in a more clear way. The resources, that were specified, must migrate into the `filtering_policy.management_events_filter.resource_scope`

Before

```terraform
filter {
  path_filter {
    any_filter {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
    }
  }
}
```

After

```terraform
filtering_policy {
  management_events_filter {
    resource_scope {
      resource_id   = "home-folder"
      resource_type = "resource-manager.folder"
    }
  }
}
```
