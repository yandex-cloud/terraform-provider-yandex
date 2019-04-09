## 0.4.0 (Unreleased)
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
* provider: authentication with service account key file. ([#3](https://github.com/yandex-cloud/terraform-provider-yandex/issues/3))
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
