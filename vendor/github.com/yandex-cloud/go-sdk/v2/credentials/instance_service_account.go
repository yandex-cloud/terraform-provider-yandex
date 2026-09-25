package credentials

// InstanceServiceAccount returns credentials for Compute Instance Service Account.
// That is, for SDK build with InstanceServiceAccount credentials and used on Compute Instance
// created with yandex.cloud.compute.v1.CreateInstanceRequest.service_account_id, API calls
// will be authenticated with this ServiceAccount ID.
// You can override the default address of Metadata Service by setting env variable.
// https://yandex.cloud/ru/docs/compute/operations/vm-control/vm-connect-sa#cli_1
func InstanceServiceAccount() NonExchangeableCredentials {
	return MetadataService()
}
