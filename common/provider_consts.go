package common

const (
	DefaultMaxRetries      = 5
	DefaultEndpoint        = "api.cloud.yandex.net:443"
	DefaultStorageEndpoint = "storage.yandexcloud.net"
	DefaultYMQEndpoint     = "message-queue.api.cloud.yandex.net"
	DefaultRegion          = "ru-central1"
)

var Descriptions = map[string]string{
	"endpoint": "The endpoint for API calls, default value is **" + DefaultEndpoint + "**.\n" +
		"This can also be defined by environment variable `YC_ENDPOINT`.",

	"folder_id": "The ID of the [Folder](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#folder) to operate under, if not specified by a given resource.\n" +
		"This can also be specified using environment variable `YC_FOLDER_ID`.",

	"cloud_id": "The ID of the [Cloud](https://yandex.cloud/docs/resource-manager/concepts/resources-hierarchy#cloud) to apply any resources to.\n" +
		"This can also be specified using environment variable `YC_CLOUD_ID`.",

	"region_id": "[The region](https://yandex.cloud/docs/overview/concepts/region) where operations will take place. For example `ru-central1`.",

	"zone": "The default [availability zone](https://yandex.cloud/docs/overview/concepts/geo-scope) to operate under, if not specified by a given resource.\n" +
		"This can also be specified using environment variable `YC_ZONE`.",

	"token": "Security token or IAM token used for authentication in Yandex Cloud.\n" +
		"Check [documentation](https://yandex.cloud/docs/iam/operations/iam-token/create) about how to create IAM token. This can also be specified using environment variable `YC_TOKEN`.",

	"service_account_key_file": "Contains either a path to or the contents of the [Service Account file](https://yandex.cloud/docs/iam/concepts/authorization/key) in JSON format.\n" +
		"This can also be specified using environment variable `YC_SERVICE_ACCOUNT_KEY_FILE`. You can read how to create service account key file [here](https://yandex.cloud/docs/iam/operations/iam-token/create-for-sa#keys-create).\n\n" +
		"~> Only one of `token` or `service_account_key_file` must be specified.\n\n" +
		"~> One can authenticate via instance service account from inside a compute instance. In order to use this method, omit both `token`/`service_account_key_file` and attach service account to the instance. [Working with Yandex Cloud from inside an instance](https://yandex.cloud/docs/compute/operations/vm-connect/auth-inside-vm).\n\n",

	"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted, default value is `false`.",

	"plaintext": "Disable use of TLS. Default value is `false`.",

	"max_retries": "This is the maximum number of times an API call is retried, in the case where requests are being throttled or experiencing transient failures. The delay between the subsequent API calls increases exponentially.",

	"storage_endpoint": "Yandex Cloud [Object Storage Endpoint](https://yandex.cloud/docs/storage/s3/#request-url), which is used to connect to `S3 API`. Default value is **" + DefaultStorageEndpoint + "**.",

	"storage_access_key": "Yandex Cloud Object Storage access key, which is used when a storage data/resource doesn't have an access key explicitly specified. \n" +
		"This can also be specified using environment variable `YC_STORAGE_ACCESS_KEY`.",

	"storage_secret_key": "Yandex Cloud Object Storage secret key, which is used when a storage data/resource doesn't have a secret key explicitly specified.\n" +
		"This can also be specified using environment variable `YC_STORAGE_SECRET_KEY`.",

	"ymq_endpoint": "Yandex Cloud Message Queue service endpoint. Default value is **" + DefaultYMQEndpoint + "**.",

	"ymq_access_key": "Yandex Cloud Message Queue service access key, which is used when a YMQ queue resource doesn't have an access key explicitly specified.\n" +
		"  This can also be specified using environment variable `YC_MESSAGE_QUEUE_ACCESS_KEY`.",

	"ymq_secret_key": "Yandex Cloud Message Queue service secret key, which is used when a YMQ queue resource doesn't have a secret key explicitly specified.\n" +
		"This can also be specified using environment variable `YC_MESSAGE_QUEUE_SECRET_KEY`.",

	"shared_credentials_file": "Shared credentials file path.\nSupported keys: `storage_access_key` and `storage_secret_key`.\n\n" +
		"~> The `storage_access_key` and `storage_secret_key` attributes from the shared credentials file are used only when the provider and a storage data/resource do not have an access/secret keys explicitly specified.\n",

	"profile": "Profile name to use in the shared credentials file. Default value is `default`.",

	"organization_id": "The ID of the [Cloud Organization](https://yandex.cloud/docs/organization/quickstart) to operate under.",
}
