---
layout: "yandex"
page_title: "Provider: Yandex.Cloud"
sidebar_current: "docs-yandex-index"
description: |-
  The Yandex.Cloud provider is used to interact with Yandex.Cloud services.
  The provider needs to be configured with the proper credentials before it can be used.
---

# Yandex.Cloud Provider

The Yandex.Cloud provider is used to interact with
[Yandex.Cloud services](https://cloud.yandex.com/). The provider needs
to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
// Configure the Yandex.Cloud provider
provider "yandex" {
  token                    = "auth_token_here"
  service_account_key_file = "path_to_service_account_key_file"
  cloud_id                 = "cloud_id_here"
  folder_id                = "folder_id_here"
  zone                     = "ru-central1-a"
}

// Create a new instance
resource "yandex_compute_instance" "default" {
  ...
}
```

## Configuration Reference

The following keys can be used to configure the provider.

* `token` - (Optional) Security token or IAM token used for authentication in Yandex.Cloud.

  Check [documentation][yandex-iam-create-token] about how to create IAM token.
  This can also be specified using environment variable `YC_TOKEN`.

* `service_account_key_file` - (Optional) Contains either a path to or the contents of the Service Account file in JSON format.

  This can also be specified using environment variable `YC_SERVICE_ACCOUNT_KEY_FILE`.
  You can read how to create service account key file [here][yandex-service-account-key].

~> **NOTE:** Only one of `token` or `service_account_key_file` must be specified.

~> **NOTE:** One can authenticate via instance service account from inside a compute instance. In order to use this method, omit both `token`/`service_account_key_file` and attach service account to the instance.
[Working with Yandex.Cloud from inside an instance][instance-service-account]

* `cloud_id` - (Required) The ID of the [cloud][yandex-cloud] to apply any resources to.

  This can also be specified using environment variable `YC_CLOUD_ID`.

* `folder_id` - (Required) The ID of the [folder][yandex-folder] to operate under, if not specified by a given resource.

  This can also be specified using environment variable `YC_FOLDER_ID`.

* `zone` - (Optional) The default [availability zone][yandex-zone] to operate under, if not specified by a given resource.

  This can also be specified using environment variable `YC_ZONE`.

* `endpoint` - (Optional) The endpoint for API calls, default value is api.cloud.yandex.net:443.

  This can also be defined by environment variable `YC_ENDPOINT`

* `max_retries` - (Optional) This is the maximum number of times an API call is retried, in the case where requests
  are being throttled or experiencing transient failures. The delay between the subsequent API calls increases
  exponentially.

* `storage_endpoint` — (Optional) Yandex.Cloud object storage [endpoint][yandex-storage-endpoint], which is used to connect to `S3 API`. Default value is `"storage.yandexcloud.net"`

* `storage_access_key` - (Optional) Yandex.Cloud storage service access key, which is used when a storage data/resource doesn't have an access key explicitly specified.

  This can also be specified using environment variable `YC_STORAGE_ACCESS_KEY`.

* `storage_secret_key` - (Optional) Yandex.Cloud storage service secret key, which is used when a storage data/resource doesn't have a secret key explicitly specified.

  This can also be specified using environment variable `YC_STORAGE_SECRET_KEY`.

* `ymq_access_key` - (Optional) Yandex.Cloud Message Queue service access key, which is used when a YMQ queue resource doesn't have an access key explicitly specified.

  This can also be specified using environment variable `YC_MESSAGE_QUEUE_ACCESS_KEY`.

* `ymq_secret_key` - (Optional) Yandex.Cloud Message Queue service secret key, which is used when a YMQ queue resource doesn't have a secret key explicitly specified.

  This can also be specified using environment variable `YC_MESSAGE_QUEUE_SECRET_KEY`.

[yandex-cloud]: https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#cloud
[yandex-folder]: https://cloud.yandex.com/docs/resource-manager/concepts/resources-hierarchy#folder
[yandex-zone]: https://cloud.yandex.com/docs/overview/concepts/geo-scope
[yandex-service-account-key]: https://cloud.yandex.com/docs/iam/operations/iam-token/create-for-sa#keys-create
[instance-service-account]: https://cloud.yandex.com/docs/compute/operations/vm-connect/auth-inside-vm
[yandex-iam-create-token]: https://cloud.yandex.com/docs/iam/operations/iam-token/create
[yandex-storage-endpoint]: https://cloud.yandex.com/en-ru/docs/storage/s3/#request-url
