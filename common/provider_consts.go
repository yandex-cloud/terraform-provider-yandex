package common

const (
	DefaultMaxRetries      = 5
	DefaultEndpoint        = "api.cloud.yandex.net:443"
	DefaultStorageEndpoint = "storage.yandexcloud.net"
	DefaultYMQEndpoint     = "message-queue.api.cloud.yandex.net"
	DefaultRegion          = "ru-central1"
)

var Descriptions = map[string]string{
	"endpoint": "The API endpoint for Yandex.Cloud SDK client.",

	"folder_id": "The default folder ID where resources will be placed.",

	"cloud_id": "ID of Yandex.Cloud tenant.",

	"region_id": "The region where operations will take place. Examples\n" +
		"are ru-central1",

	"zone": "The zone where operations will take place. Examples\n" +
		"are ru-central1-a, ru-central2-c, etc.",

	"token": "The access token for API operations.",

	"service_account_key_file": "Either the path to or the contents of a Service Account key file in JSON format.",

	"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
		"default value is `false`.",

	"plaintext": "Disable use of TLS. Default value is `false`.",

	"max_retries": "The maximum number of times an API request is being executed. \n" +
		"If the API request still fails, an error is thrown.",

	"storage_endpoint": "Yandex.Cloud storage service endpoint. Default is \n" + DefaultStorageEndpoint,

	"storage_access_key": "Yandex.Cloud storage service access key. \n" +
		"Used when a storage data/resource doesn't have an access key explicitly specified.",

	"storage_secret_key": "Yandex.Cloud storage service secret key. \n" +
		"Used when a storage data/resource doesn't have a secret key explicitly specified.",

	"ymq_endpoint": "Yandex.Cloud Message Queue service endpoint. Default is \n" + DefaultYMQEndpoint,

	"ymq_access_key": "Yandex.Cloud Message Queue service access key. \n" +
		"Used when a message queue resource doesn't have an access key explicitly specified.",

	"ymq_secret_key": "Yandex.Cloud Message Queue service secret key. \n" +
		"Used when a message queue resource doesn't have a secret key explicitly specified.",

	"shared_credentials_file": "Path to shared credentials file.",

	"profile": "Profile to use in the shared credentials file. Default value is `default`.",
}
