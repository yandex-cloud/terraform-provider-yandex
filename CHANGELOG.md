## 0.29.0 (Unreleased)
FEATURES:
* **New Data Source:** `yandex_mdb_postgresql_cluster`
* **New Resource:** `yandex_mdb_postgresql_cluster`
* **New Data Source:** `yandex_mdb_mysql_cluster`
* **New Resource:** `yandex_mdb_mysql_cluster`

## 0.28.0 (January 13, 2020)
FEATURES:
* **New Resource:** `yandex_kms_symmetric_key`

ENHANCEMENTS:
* managed-kubernetes: support maintenance policy for `yandex_kubernetes_cluster` and `yandex_kubernetes_node_group`
* lb: `yandex_lb_network_load_balancer` now supports UDP for `protocol` attribute.

BUG FIXES:
* provider: fix `message type "google.protobuf.Empty" isn't linked in` ([#35](https://github.com/terraform-providers/terraform-provider-yandex/issues/35))

## 0.27.0 (December 20, 2019)
FEATURES:
* **New Data Source:** `yandex_mdb_mongodb_cluster`
* **New Resource:** `yandex_mdb_mongodb_cluster`

ENHANCEMENTS:
* mdb: support sharding in `yandex_mdb_clickhouse_cluster`
* lb: changes in `yandex_lb_network_load_balancer` support ipv6 at listener.0.external_address_spec.ip_version

## 0.26.0 (December 06, 2019)
ENHANCEMENTS:
* compute: support for custom rules in `yandex_compute_instance_group.scale_policy.auto_scale`

## 0.25.0 (December 05, 2019)
FEATURES:
* **New Data Source:** `yandex_mdb_clickhouse_cluster`
* **New Resource:** `yandex_mdb_clickhouse_cluster`

## 0.24.0 (December 03, 2019)
BUG FIXES:
* managed-kubernetes: changes in `yandex_kubernetes_node_group` allocation_policy should trigger destroy/add.
* managed-kubernetes: changes in `yandex_kubernetes_cluster` location, release_channel should trigger destroy/add.
* managed-kubernetes: changes in `yandex_kubernetes_cluster` master.0.version should NOT trigger destroy/add, and use update instead.
* managed-kubernetes: forbidden zero values in `yandex_kubernetes_node_group`, in instance_template.0.resources.0.memory(cores)
* managed-kubernetes: fill `instance_group_id` field in `yandex_kubernetes_node_group` datasource and resource.

ENHANCEMENTS:
* compute: support update of service_account_id in `yandex_compute_instance` without resource recreation.
* datasource resolving by name now uses folder_id from its config (when provided), affected datasources: 
`yandex_compute_disk`,  `yandex_compute_image`, `yandex_compute_instance`, `yandex_compute_snapshot`,
`yandex_container_registry`, `yandex_kubernetes_cluster`, `yandex_kubernetes_node_group`,
`yandex_lb_network_load_balancer`, `yandex_lb_target_group`, `yandex_mdb_redis_cluster`,
`yandex_vpc_network`, `yandex_vpc_route_table`, `yandex_vpc_subnet`.

## 0.23.0 (November 05, 2019)
ENHANCEMENTS:
* mdb: support sharding in `yandex_mdb_redis_cluster`
* compute: increase `yandex_compute_snapshot` timeout from 5 to 20 minutes

BUG FIXES:
* managed-kubernetes: mark as computable `version` and `public_ip` in `yandex_kubernetes_cluster` resource

## 0.22.0 (October 24, 2019)
ENHANCEMENTS:
* compute: add `instances` to `yandex_compute_instance_group` resource
* mdb: add fqdns of hosts in `yandex_mdb_redis_cluster` resource and data source
* managed-kubernetes: add `version` to `yandex_kubernetes_node_group` resource

## 0.21.0 (October 17, 2019)
ENHANCEMENTS:
* storage: `yandex_storage_bucket` and `yandex_storage_object` resources can manage ACL

## 0.20.0 (October 15, 2019)
FEATURES:
* **New Resource:** `yandex_storage_bucket`
* **New Resource:** `yandex_storage_object`

## 0.19.0 (October 15, 2019)
ENHANCEMENTS:
* managed-kubernetes: `yandex_kubernetes_node_group` resource can now be imported
* managed-kubernetes: `yandex_kubernetes_cluster` resource can now be imported

BUG FIXES:
* minor documentation fixes for Kubernetes cluster resource and instance group datasource.

## 0.18.0 (October 11, 2019)
ENHANCEMENTS:
* provider: support authentication via instance service account from inside an instance

BUG FIXES:
* container: increase default timeout

## 0.17.0 (October 02, 2019)
FEATURES:
* compute: auto_scale support added for `yandex_compute_instance_group` resource and data source

## 0.16.0 (October 01, 2019)
* **New Data Source:** `yandex_mdb_redis_cluster`
* **New Resource:** `yandex_mdb_redis_cluster`

## 0.15.0 (September 30, 2019)
FEATURES:
* **New Data Source:** `yandex_kubernetes_cluster`
* **New Data Source:** `yandex_kubernetes_node_group`
* **New Resource:** `yandex_kubernetes_cluster`
* **New Resource:** `yandex_kubernetes_node_group`

## 0.14.0 (September 27, 2019)
* provider: migrate to standalone Terraform SDK module ([#22](https://github.com/terraform-providers/terraform-provider-yandex/issues/22))
* provider: support graceful shutdown
* iam: use logic lock on cloud while create SA to prevent simultaneous IAM membership changes
* container: resolve data source `yandex_container_registry` by name.

## 0.13.0 (September 23, 2019)
FEATURES:
* **New Resource:** `yandex_iam_service_account_api_key`
* **New Resource:** `yandex_iam_service_account_key`

ENHANCEMENTS:
* compute: `yandex_compute_snapshot` resource can now be imported
* iam: `yandex_iam_service_account` resource can now be imported
* iam: `yandex_iam_service_account_static_access_key` resource now supports `pgp_key` field.

## 0.12.0 (September 20, 2019)
FEATURES:
* **New Data Source:** `yandex_container_registry`
* **New Resource:** `yandex_container_registry`

## 0.11.2 (September 19, 2019)
ENHANCEMENTS:
* provider: provider uses permanent client-request-id identifier while the terraform is running

BUG FIXES:
* provider: fix provider name and version detection   

## 0.11.1 (September 13, 2019)
ENHANCEMENTS:
* provider: set provider name and version in user agent header.

BUG FIXES:
* compute: fix flattening of health checks for `yandex_compute_instance_group` resource   

## 0.11.0 (September 11, 2019)
ENHANCEMENTS:
* compute: add `resources.0.gpus` attribute in `yandex_compute_instance` resource and data source 
* compute: add `resources.0.gpus` attribute in `yandex_compute_instance_group` resource and data source

## 0.10.2 (September 09, 2019)
ENHANCEMENTS:
* compute: `yandex_compute_snapshot` resource can now be imported
* iam: `yandex_iam_service_account` resource can now be imported

BUG FIXES:
* compute: fix read operation of `yandex_compute_instance`  

## 0.10.1 (August 26, 2019)
BUG FIXES:
* resourcemanager: resources `yandex_*_iam_binding`, `yandex_•_iam_policy` works with full set of bindings.

## 0.10.0 (August 21, 2019)
BUG FIXES:
* vpc: remove `v6_cidr_blocks` attr in `yandex_vpc_subnet` resource. This property is not available right now.

ENHANCEMENTS:
* compute: instance_group data source and resource support new fields in `load_balancer` section.
* resourcemanager: support lookup `yandex_resourcemanager_folder` at specific cloud_id. ([#17](https://github.com/terraform-providers/terraform-provider-yandex/issues/17))

## 0.9.1 (August 14, 2019)
ENHANCEMENTS:
* compute: use `min_disk_size` of image or `disk_size` of snapshot to set size of boot_disk on instance template for `yandex_compute_instance_group`.

## 0.9.0 (August 07, 2019)
FEATURES:
* **New Data Source:** `yandex_lb_network_load_balancer`
* **New Data Source:** `yandex_lb_target_group`
* **New Resource:** `yandex_lb_network_load_balancer`
* **New Resource:** `yandex_lb_target_group`

ENHANCEMENTS:
* compute: use `min_disk_size` of image or `disk_size` of snapshot to set size of boot_disk on instance create.
* compute: update instance resource spec and platform type in one request.

BUG FIXES:
* compute: change attribute `folder_id` from Required to Optional for `yandex_compute_instance_group` resource [[#14](https://github.com/terraform-providers/terraform-provider-yandex/issues/14)].
   
## 0.8.1 (July 04, 2019)
BUG FIXES:
* compute: fix `yandex_compute_instance_group` with `load_balancer_spec` defined [[#13](https://github.com/terraform-providers/terraform-provider-yandex/issues/13)].   

## 0.8.0 (June 25, 2019)
FEATURES:
* **New Data Source**: `yandex_compute_instance_group`
* **New Resource**: `yandex_compute_instance_group`

## 0.7.0 (June 06, 2019)
ENHANCEMENTS:
* provider: Support SDK retries.  
 
## 0.6.0 (May 29, 2019)
NOTES:
* provider: This release includes a Terraform SDK upgrade with compatibility for Terraform v0.12. 
* provider: Switch dependency management to Go modules. ([#5](https://github.com/terraform-providers/terraform-provider-yandex/issues/5))

## 0.5.2 (April 24, 2019)
ENHANCEMENTS:
* compute: fractional values for memory for `yandex_compute_instance`.
* compute: `yandex_compute_instance` support update platform_id in stopped state.

## 0.5.1 (April 20, 2019)
BUG FIXES:
* compute: fix migration process for `yandex_compute_instance`.   

## 0.5.0 (April 17, 2019)
ENHANCEMENTS:
* all: save new entity identifiers at start of create operation
* compute: `yandex_compute_instance` support update resources in stopped state.
* compute: change attribute `resources` type from Set to List

## 0.4.1 (April 11, 2019)
BUG FIXES:
* compute: fix properties of `service_account_id` attribute.   

## 0.4.0 (April 09, 2019)
ENHANCEMENTS:
* compute: `yandex_compute_instance` adds a `service_account_id` attribute.

## 0.3.0 (April 03, 2019)
FEATURES:
* **New Datasource**: `yandex_vpc_route_table`
* **New Resource**: `yandex_vpc_route_table` 

ENHANCEMENTS:
* vpc: `yandex_vpc_subnet` adds a `route_table_id` field.

## 0.2.0 (March 26, 2019)
ENHANCEMENTS:
* provider: authentication with service account key file. ([#3](https://github.com/terraform-providers/terraform-provider-yandex/issues/3))
* vpc: increase subnet create/update/delete timeout.
* vpc: resolve data source `network`, `subnet` by name.
* compute: resolve data source `instance`, `disk`, `image`, `snapshot` objects by names.
* resourcemanager: resolve data source `folder` by name.

## 0.1.16 (March 14, 2019)
ENHANCEMENTS:
* compute: support preemptible instance type.   

BUG FIXES:
* compute: fix update method on compute resources for description attribute.
   
## 0.1.15 (February 22, 2019)

BACKWARDS INCOMPATIBILITIES:
* compute: `yandex_compute_disk.source_image_id` and `yandex_compute_disk.source_snapshot_id` has been removed.
* iam: `iam_service_account_key` was renamed to `iam_service_account_static_access_key`.

ENHANCEMENTS:
* provider: more descriptive error messages.
* compute: `yandex_compute_disk` support for increasing size without force recreation of the resource.   

BUG FIXES:
* compute: make consistent disk type attribute name `type_id` -> `type`.   
* compute: remove attr `instance_id` from `yandex_compute_instance`.
* compute: make `yandex_compute_instancenet.network_interface.*.nat` ForceNew.

## 0.1.14 (December 26, 2018)

FEATURES:
* **New Data Source:** `yandex_compute_disk`
* **New Data Source:** `yandex_compute_image`
* **New Data Source:** `yandex_compute_instance`
* **New Data Source:** `yandex_compute_snapshot`
* **New Data Source:** `yandex_iam_policy`
* **New Data Source:** `yandex_iam_role`
* **New Data Source:** `yandex_iam_service_account`
* **New Data Source:** `yandex_iam_user`
* **New Data Source:** `yandex_resourcemanager_cloud`
* **New Data Source:** `yandex_resourcemanager_folder`
* **New Data Source:** `yandex_vpc_network`
* **New Data Source:** `yandex_vpc_subnet`
* **New Resource:** `yandex_compute_disk`
* **New Resource:** `yandex_compute_image`
* **New Resource:** `yandex_compute_instance`
* **New Resource:** `yandex_compute_snapshot`
* **New Resource:** `yandex_iam_service_account`
* **New Resource:** `yandex_iam_service_account_iam_binding`
* **New Resource:** `yandex_iam_service_account_iam_member`
* **New Resource:** `yandex_iam_service_account_iam_policy`
* **New Resource:** `yandex_iam_service_account_key`
* **New Resource:** `yandex_vpc_network`
* **New Resource:** `yandex_vpc_subnet`
* **New Resource:** `yandex_resourcemanager_cloud_iam_binding`
* **New Resource:** `yandex_resourcemanager_cloud_iam_member`
* **New Resource:** `yandex_resourcemanager_folder_iam_binding`
* **New Resource:** `yandex_resourcemanager_folder_iam_member`
* **New Resource:** `yandex_resourcemanager_folder_iam_policy`

ENHANCEMENTS:
* compute: support IPv6 addresses
* vpc: support IPv6 addresses
