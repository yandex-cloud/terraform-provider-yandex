## 0.57.0 (Unreleased)
## 0.56.0 (April 15, 2021)
ENHANCEMENTS:
* dataproc: supported `security_group_ids`
* add `dns_record`, `ipv6_dns_record` and `nat_dns_record` to `network_interface` entity in `yandex_compute_instance_group`
* ydb: support for Yandex Database clusters
* compute: increase disk size limit from 4096Gb to 8192Gb 
* vpc: add `name` field description at vpc_security_group datasource and example of it usage

BUG FIXES:
* compute: placement_policy update in `yandex_compute_instance_group`

## 0.55.0 (April 1, 2021)
FEATURES:
* storage: `yandex_storage_bucket` resource supports bucket policy configuration

ENHANCEMENTS:
* add extended API logging. Use TF_ENABLE_API_LOGGING=1 with TF_LOG=debug to see extended output.
* support IAM token in tests
* managed-redis: added 'tls_enabled' field
* managed-kafka: added 'unmanaged_topics' cluster flag and some cluster config flags
* mdb: add `host` attribute in `yandex_mdb_kafka_cluster` resource and data source

BUG FIXES:
* serverless: fix API Gateway specification update

## 0.54.0 (March 23, 2021)
ENHANCEMENTS:
* provider: the default development, testing and building of the provider is now done with Go 1.16.
* serverless: supported API Gateway

BUG FIXES:
* mdb: fix user settings diff for ClickHouse cluster

## 0.53.0 (March 19, 2021)
ENHANCEMENTS:
* mdb: add example and update documentation for `yandex_mdb_postgresql_cluster` resource
* serverless: supported log-group trigger

BUG FIXES:
* dns: fix recordset update in `yandex_dns_recordset`
* storage: Fix timeout while applying CORS settings with empty fields

## 0.52.0 (March 10, 2021)
FEATURES:
* **New Resource:** `yandex_mdb_sqlserver_cluster`
* managed-kubernetes: support `security_group_ids` for `yandex_kubernetes_node_group` resource and data source
* **New Resource:** `yandex_dns_zone`
* **New Resource:** `yandex_dns_recordset`
* **New Data Source:** `yandex_dns_zone`
* serverless: support import for all resources
* **New Resource:** `yandex_container_repository`
* **New Resource:** `yandex_container_repository_iam_binding`
* **New Data Source:** `yandex_container_repository`

ENHANCEMENTS:
* mdb: add `service_account_id` section in `yandex_mdb_clickhouse_cluster` resource and data source
* mdb: add `cloud_storage` section in `yandex_mdb_clickhouse_cluster` resource and data source
* managed-kubernetes: added `network_interface` section for `yandex_kubernetes_node_group`
* managed-redis: added 'disk_type_id' field

WARNING:
* managed-kubernetes: `nat` entitiy and `subnet_id` entity in `allocation_policy` section for `yandex_kubernetes_node_group` is now deprecated
* mdb: when changing the `assign_public_ip` attribute to `host` entity in  `yandex_mdb_mysql_cluster`, the old host is deleted and a new host is created
* mdb: add `allow_regeneration_host` attribute in `yandex_mdb_mysql_cluster` resource

BUG FIXES:
* mdb: fix host delete in `yandex_mdb_mysql_cluster`

## 0.51.1 (February 20, 2021)
ENHANCEMENTS:
* compute: add documentation and example for non-replicated disk

## 0.51.0 (February 19, 2021)
FEATURES:
* compute: support yandex_disk_placement_group resource and data source.
* compute: integrate yandex_disk_placement_group with compute disk resource and data source.
* mdb: added the ability to upgrade the Mysql version using the `version` attribute in `yandex_mdb_mysql_cluster`

ENHANCEMENTS:
* mdb: add `restore` entity in `yandex_mdb_mysql_cluster` resource
* mdb: add `connection_limits`, `global_permissions` and `authentication_plugin` attributes to `user` entity in `yandex_mdb_mysql_cluster` resource and data source
* mdb: add `restore` entity in `yandex_mdb_postgresql_cluster` resource
* mdb: add `settings` and `quota` sections to `user` entity in `yandex_mdb_clickhouse_cluster` resource and data source.
* iam: corrected documentation for `yandex_resourcemanager_cloud_iam_binding` resource.
* iam: corrected documentation for `yandex_resourcemanager_folder_iam_binding` resource.

BUG FIXES:
* mdb: fix updating user permissions for Kafka cluster

WARNING:
* mdb: replace sets with lists for users in `yandex_mdb_mysql_cluster`. There can appear changes in diff for users, which will not change anything and will disappear after apply

## 0.50.0 (February 5, 2021)
FEATURES:
* **New Resource:** `yandex_container_registry_iam_binding`
* mdb: version 13 is available in `yandex_mdb_postgresql_cluster`
* storage: `yandex_storage_bucket` resource supports versioning configuration
* storage: `yandex_storage_bucket` resource supports logging configuration
* vpc: add example for ddos protected address documentation
* compute: support yandex_placement_group resource and data source.
* compute: integrate yandex_placement_group with compute instance and instance group resources and data source.

ENHANCEMENTS:
* storage: add bucket configuration example
* mdb: support `security_group_ids` for managed service for kafka
* mdb: add `web_sql` and `data_lens` attribute to `access` entity in `yandex_mdb_mysql_cluster` resource and data source
* mdb: add `mysql_config` attribute in `yandex_mdb_mysql_cluster` resource and data source
* mdb: add `format_schema` section in `yandex_mdb_clickhouse_cluster` resource and data source
* mdb: add `ml_model` section in `yandex_mdb_clickhouse_cluster` resource and data source
* mdb: add `replication_source_name`, `priority` attributes to `host`entity and `host_master_name` attribute in `yandex_mdb_postgresql_cluster` resource and data source
* mdb: add `sql_user_management` and `sql_database_management` attributes in `yandex_mdb_clickhouse_cluster` resource and data_source
* mdb: add `admin_password` attribute in `yandex_mdb_clickhouse_cluster` resource
* kms: add sensitive flag for `plaintext` attribute in `yandex_kms_secret_ciphertext` resource
* managed-kubernetes: support `security_group_ids` for `yandex_kubernetes_cluster` resource and data source

## 0.49.0 (January 15, 2021)
FEATURES:
* storage: `yandex_storage_bucket` resource supports lifecycle configuration

ENHANCEMENTS:
* mdb: changing `folder_id` attribute in `yandex_mdb_postgresql_cluster` moves PostgreSQL cluster to new folder
* mdb: add `web_sql` attribute to `config.access` entity in `yandex_mdb_postgresql_cluster` resource and data source
* mdb: add `performance_diagnostics` attribute to `config` entity in `yandex_mdb_postgresql_cluster` resource and data source
* mdb: add `settings` attribute to `user` entity in `yandex_mdb_postgresql_cluster` resource and data source
* mdb: add `postgresql_config` attribute to `config` entity in `yandex_mdb_postgresql_cluster` resource and data
* mdb: support `security_group_ids` in all database cluster resources and data sources
* compute: `strategy` attribute to `deploy_policy` entity in `yandex_compute_instance_group` resource and data source
* vpc: extend validation for listener spec in `yandex_lb_network_load_balancer` resource

## 0.48.0 (December 22, 2020)
BUG FIXES:
* mdb: fix setting of folder_id field for MongoDB cluster
* dataproc: add documentation for the `ui_proxy` attribute
* vpc: fix panic on reading `yandex_vpc_address` resource

ENHANCEMENTS:
* mdb: add `conn_limit` attribute to `user` entity in `yandex_mdb_postgresql_cluster` resource and data source
* mdb: add `config` section in `yandex_mdb_clickhouse_cluster` resource and data source

## 0.47.0 (November 10, 2020)
BUG FIXES:
* kms: fix import operation
* serverless: folder_id can be using from yandex_function, yandex_function_trigger, yandex_iot_core_registry
* serverless: crash fix for dlq option in yandex_function_trigger

ENHANCEMENTS:
* vpc: default_security_group_id field was added to network resource and data source
* provider: support authentication via IAM token

FEATURES:
* mdb: support ClickHouse shard groups in `yandex_mdb_clickhouse_cluster`

## 0.46.0 (October 19, 2020)
BUG FIXES:
* vpc: Security group rule port bugfix: can create rules without specifying a port
* vpc: Fix internal_address_spec block in network load balancer resource doc
* vpc: Security group ANY port bug fix
* dataproc: support for UI Proxy

ENHANCEMENTS:
* serverless: improved zip archive content size limit excession error message

## 0.45.1 (October 06, 2020)

BUG FIXES:
* fix release issue

## 0.45.0 (October 05, 2020)
FEATURES:
* mdb: support MongoDB 4.4 in `yandex_mdb_mongodb_cluster`
* vpc: address resource & data source

ENHANCEMENTS:
* lb: improve NLB sweeper and tests

BUG FIXES:
* vpc: `static_route` in `yandex_vpc_route_table` is optional now

## 0.44.1 (September 24, 2020)

BUG FIXES:
* vpc: fix "security_group" data source

## 0.44.0 (September 22, 2020)

FEATURES:
* vpc: security group rule targets `security_group_id` and `predefined_target` are supported
* storage: `yandex_storage_bucket` resource can manage SSE

ENHANCEMENTS:
* some changes in security group resource

BUG FIXES:
* lb: fix modifying listener settings

## 0.43.0 (August 20, 2020)

FEATURES:
* iam: support for resolving by name in `yandex_iam_service_account` data source

BUG FIXES:
* compute: fix `yandex_compute_instance` update trying to re-configure dymanic NAT
* mdb: replace sets with lists for users and databases in `yandex_mdb_postgresql_cluster`.
WARNING: there can appear changes in diff for users and databases, which will not change anything and will disappear after apply

## 0.42.1 (August 04, 2020)

BUG FIXES:
* compute: fix panic on parsing `instance_template.network_interface.security_group_ids` attribute in `yandex_compute_instance_group` resource

## 0.42.0 (July 27, 2020)
FEATURES:
* mdb: support Redis 6.0 in `yandex_mdb_redis_cluster`

FEATURES:
* **New Data Source:** `yandex_client_config`

ENHANCEMENTS:
* mdb: add `role` attribute to `host` entity in `yandex_mdb_postgresql_cluster` resource and data source
* compute: support update of `network_interface` attribute for `yandex_compute_instance` resource

BUG FIXES:
* compute: fix `secondary_disk` validation in `yandex_compute_instance_group` resource
* compute: remove `secondary_disk.security_group_ids` attribute from `yandex_compute_instance` data source

## 0.41.1 (June 24, 2020)

BUG FIXES:
* vpc: fix panic on empty subnet dhcp options on `yandex_vpc_subnet` resource ([#82](https://github.com/terraform-providers/terraform-provider-yandex/issues/82))

## 0.41.0 (June 23, 2020)
FEATURES:
* **New Data Source:** `yandex_message_queue`
* **New Resource:** `yandex_message_queue`
* vpc: allow setting dhcp options for `yandex_vpc_subnet`

ENHANCEMENTS:
* mdb: document mdb enumerables for Redis, MongoDB and ClickHouse
* provider: support set 'service_account_key_file' as either the path to or the contents of a Service Account key file in JSON format
* managed-kubernetes: support `gpus` attribute for `yandex_kubernetes_node_group`
* compute: add `instance_template.scale_policy.test_auto_scale` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `deletion_protection` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `instance_template.network_interface.security_group_ids` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `network_interface.security_group_ids` attribute in `yandex_compute_instance` resource and data source

BUG FIXES:
* mdb: fix typo in using mdb api by `resource_yandex_mdb_postgresql_cluster`
* vpc: fix removing `yandex_vpc_subnet` attribute `route_table_id`

## 0.40.0 (May 22, 2020)
FEATURES:
* **New Resource:** `yandex_function_iam_binding`

BUG FIXES:
* compute: add `ipv4` flag in `yandex_compute_instance` data source
* mdb: fix disk size change on `mdb_mongodb` resource update
* mdb: adding database with its owner to existing `resource_yandex_mdb_postgresql_cluster` simultaneously

## 0.39.0 (May 05, 2020)
ENHANCEMENTS:
* mdb: add `roles` attribute to `user` entity in `mdb_mongodb` resource and data source
* compute: change allowed disk type from `network-nvme` to `network-ssd`
* compute: `ipv4` flag determines whether to assign a IPv4 address for `network_interface` in `yandex_compute_instance` and `yandex_compute_instance_group`

## 0.38.0 (April 22, 2020)
FEATURES:
* **New Data Source:** `yandex_vpc_security_group`
* **New Resource:** `yandex_vpc_security_group`

ENHANCEMENTS:
* managed-kubernetes: allow to create cluster with KMS provider for secrets encryption.

## 0.37.0 (April 16, 2020)
ENHANCEMENTS:
* storage: support custom acl grants for `yandex_storage_bucket`

## 0.36.0 (April 16, 2020)
ENHANCEMENTS:
* compute: add `variables` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `status` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `instance_template.name` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `instance_template.hostname` attribute in `yandex_compute_instance_group` resource and data source
* compute: add `instances.status_changed_at` attribute in `yandex_compute_instance_group` resource and data source
* managed-kubernetes: add `node_ipv4_cidr_mask_size` attribute to `yandex_kubernetes_cluster` resource and data source
* managed-kubernetes: add `deploy_policy.max_unavailable` and `deploy_policy.max_expansion` attributes to `yandex_kubernetes_node_group` resource and data source
* serverless: add `environment` attribute in `yandex_function` resource and data source

BUG FIXES:
* mdb: fix modifying `yandex_mdb_mysql_cluster` attribute `backup_window_start`

## 0.35.0 (March 31, 2020)
FEATURES:
* **New Resource:** `yandex_kms_secret_ciphertext`

ENHANCEMENTS:
* mdb: add `config_spec.access.serverless` attribute in `resource_yandex_mdb_clickhouse_cluster` resource and data source

BUG FIXES:
* mdb: forbidden to change user settings that are not present in the scheme of `resource_yandex_mdb_postgresql_cluster`
* compute: compute instance attribute `hostname` is now filled when imported

## 0.34.0 (March 18, 2020)
FEATURES:
* **New Data Source:** `yandex_function`
* **New Data Source:** `yandex_function_trigger`
* **New Resource:** `yandex_function`
* **New Resource:** `yandex_function_trigger`

ENHANCEMENTS:
* compute: add `instance_template.network_settings` attribute in `yandex_compute_instance_group` resource and data source

## 0.33.0 (March 06, 2020)
FEATURES:
* **New Data Source:** `yandex_iot_core_registry`
* **New Data Source:** `yandex_iot_core_device`
* **New Resource:** `yandex_iot_core_registry`
* **New Resource:** `yandex_iot_core_device`

ENHANCEMENTS:
* managed-kubernetes: support network policy provider for `yandex_kubernetes_cluster` ([#45](https://github.com/terraform-providers/terraform-provider-yandex/issues/45))
* managed-kubernetes: add `node_labels`, `node_taints` and `allowed_unsafe_sysctls` fields for `yandex_kubernetes_node_group`

BUG FIXES:
* mdb: throw error when trying to modify `assign_public_ip` in `resource_yandex_mdb_postgresql_cluster`,`resource_yandex_mdb_mysql_cluster`

## 0.32.0 (February 28, 2020)
FEATURES:
* compute: allow setting public IP address for `yandex_compute_instance`
* mdb: support ClickHouse cluster access for Yandex.Metrica

BUG FIXES:
* mdb: disallow change `lc_collate` and `lc_ctype` in `yandex_mdb_postgresql_cluster` after creation.
* container-registry: prevent terraform crash during `terraform destroy` execution for non-existent container registry.
* resourcemanager: data source folder. fixed folder_id resolve by name (would fail without folder_id in provider config)
* managed-kubernetes: k8s cluster version update no longer requires destroying of resource
* managed-kubernetes: update of maintenance window `start_time` and `duration` will now trigger resource update

## 0.31.0 (February 14, 2020)
FEATURES:
* compute: software network acceleration for `yandex_compute_instance`
* mdb: support roles for users in `yandex_mdb_mysql_cluster`

ENHANCEMENTS:
* compute: support metric labels for custom rules in `yandex_compute_instance_group` resource `scale_policy.auto_scale` attribute

BUG FIXES:
* lb: fix modifying health check settings

## 0.30.0 (February 04, 2020)
FEATURES:
* **New Data Source:** `yandex_dataproc_cluster`
* **New Resource:** `yandex_dataproc_cluster`

ENHANCEMENTS:
* managed-kubernetes: support autoscaled `yandex_kubernetes_node_group`

## 0.29.1 (January 29, 2020)
BUG FIXES:
* resourcemanager: data source cloud. fixed cloud_id resolve by name (would fail without folder_id in provider config)

## 0.29.0 (January 24, 2020)
FEATURES:
* **New Data Source:** `yandex_mdb_mysql_cluster`
* **New Data Source:** `yandex_mdb_postgresql_cluster`
* **New Resource:** `yandex_mdb_mysql_cluster`
* **New Resource:** `yandex_mdb_postgresql_cluster`

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
