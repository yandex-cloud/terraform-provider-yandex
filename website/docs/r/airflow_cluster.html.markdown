---
layout: "yandex"
page_title: "Yandex: yandex_airflow_cluster"
sidebar_current: "docs-yandex-airflow-cluster"
description: |-
  Manages an Apache Airflow cluster within Yandex.Cloud.
---

# yandex\_airflow\_cluster

Manages an Apache Airflow cluster within Yandex.Cloud. For more information, see
[the official documentation](https://yandex.cloud/docs/managed-airflow/concepts/).

## Example Usage

Example of creating an Apache Airflow cluster.

```hcl
resource "yandex_airflow_cluster" "this" {
  name = "airflow-created-with-terraform"
  subnet_ids = [yandex_vpc_subnet.a.id, yandex_vpc_subnet.b.id, yandex_vpc_subnet.d.id]
  service_account_id = yandex_iam_service_account.for-airflow.id
  admin_password = "some-strong-password"

  code_sync = {
    s3 = {
      bucket = "bucket-for-airflow-dags"
    }
  }

  webserver = {
    count = 1
    resource_preset_id = "c1-m4"
  }

  scheduler = {
    count = 1
    resource_preset_id = "c1-m4"
  }

  worker = {
    min_count = 1
    max_count = 2
    resource_preset_id = "c1-m4"
  }

  airflow_config = {
    "api" = {
      "auth_backends" = "airflow.api.auth.backend.basic_auth,airflow.api.auth.backend.session"
    }
  }

  pip_packages = ["dbt"]

  lockbox_secrets_backend = {
    enabled = true
  }

  logging = {
    enabled = true
    folder_id = var.folder_id
    min_level = "INFO"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of Apache Airflow cluster.

* `folder_id` - (Optional) The ID of the folder that the cluster belongs to. If it is not provided, the default provider folder is used.

* `service_account_id` - (Required) Service account used to access Cloud resources. For more information, see [documentation](https://yandex.cloud/docs/managed-airflow/concepts/impersonation).

* `subnet_ids` - (Required) IDs of VPC network subnets where instances of the cluster are attached.

* `admin_password` - (Required) Password that is used to log in to Apache Airflow web UI under `admin` user.

* `code_sync` - (Required) Parameters of the location and access to the code that will be executed in the cluster. The structure is documented below.

* `webserver` - (Required) Configuration of webserver instances. The structure is documented below.

* `scheduler` - (Required) Configuration of scheduler instances. The structure is documented below.

* `worker` - (Required) Configuration of worker instances. The structure is documented below.

* `triggerer` - (Optional) Configuration of triggerer instances. The structure is documented below.

* `airflow_config` - (Optional) Configuration of the Apache Airflow application itself. The value of this attribute is a two-level map. 
  Keys of top-level map are the names of [configuration sections](https://airflow.apache.org/docs/apache-airflow/stable/configurations-ref.html#airflow-configuration-options).
  Keys of inner maps are the names of configuration options within corresponding section.

* `pip_packages` - (Optional) Python packages that are installed in the cluster.

* `deb_packages` - (Optional) System packages that are installed in the cluster.

* `lockbox_secrets_backend` - (Optional) Configuration of Lockbox Secrets Backend. [See documentation](https://yandex.cloud/docs/managed-airflow/tutorials/lockbox-secrets-in-maf-cluster) for details. The structure is documented below.

* `logging` - (Optional) Cloud Logging configuration. The structure is documented below.

* `security_group_ids` - (Optional) List of security groups applied to cluster components.

* `description` - (Optional) Description of the cluster.

* `labels` - (Optional) A set of key/value label pairs to assign to the cluster.

* `deletion_protection` - (Optional) Inhibits deletion of the cluster.  Can be either `true` or `false`.


- - -

The `code_sync` block supports:

* `s3` - (Required) Currently only Object Storage (S3) is supported as the source of DAG files. The structure is documented below.

- - -

The `s3` block supports:

* `bucket` - (Required) The name of the Object Storage bucket that stores DAG files used in the cluster.

- - -

The `webserver` block supports:

* `count` - (Required) The number of webserver instances in the cluster.

* `resource_preset_id` - (Required) ID of the preset for computational resources available to an instance (CPU, memory etc.).

- - -

The `scheduler` block supports:

* `count` - (Required) The number of scheduler instances in the cluster.

* `resource_preset_id` - (Required) ID of the preset for computational resources available to an instance (CPU, memory etc.).

- - -

The `worker` block supports:

* `min_count` - (Required) The minimum number of worker instances in the cluster.

* `max_count` - (Required) The maximum number of worker instances in the cluster.

* `resource_preset_id` - (Required) ID of the preset for computational resources available to an instance (CPU, memory etc.).

- - -

The `triggerer` block supports:

* `count` - (Required) The number of triggerer instances in the cluster.

* `resource_preset_id` - (Required) ID of the preset for computational resources available to an instance (CPU, memory etc.).

- - -

The `lockbox_secrets_backend` block supports:

* `enabled` - (Required) Enables usage of Lockbox Secrets Backend.

- - -

The `logging` block supports:

* `enabled` - (Required) Enables delivery of logs generated by the Airflow components to Cloud Logging.

* `folder_id` - (Optional) Logs will be written to default log group of specified folder. Exactly one of the attributes `folder_id` and `log_group_id` should be specified.

* `log_group_id` - (Optional) Logs will be written to the specified log group. Exactly one of the attributes `folder_id` and `log_group_id` should be specified.

* `min_level` - (Optional) Minimum level of messages that will be sent to Cloud Logging. Can be either `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. If not set then server default is applied (currently `INFO`).


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the cluster.

* `created_at` - Timestamp of cluster creation.

* `health` - Aggregated health of the cluster. Can be either `ALIVE`, `DEGRADED`, `DEAD` or `HEALTH_UNKNOWN`.
  For more information see `health` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).

* `status` - Status of the cluster. Can be either `CREATING`, `STARTING`, `RUNNING`, `UPDATING`, `STOPPING`, `STOPPED`, `ERROR` or `STATUS_UNKNOWN`.
  For more information see `status` field of JSON representation in [the official documentation](https://yandex.cloud/docs/managed-airflow/api-ref/Cluster/).


## Import

A cluster can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_airflow_cluster.this cluster_id
```
