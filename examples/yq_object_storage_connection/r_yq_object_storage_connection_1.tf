//
// Create a new Object Storage connection.
//

resource "yandex_yq_object_storage_connection" "my_os_connection" {
    name = "tf-test-os-connection"
    description = "Connection has been created from Terraform"
    bucket = "some-public-bucket"
    service_account_id = yandex_iam_service_account.for-yq.id
}
