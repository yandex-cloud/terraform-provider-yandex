//
// Get information about existing Datatransfer Endpoint
//
data "yandex_datatransfer_transfer" "pgpg_transfer_ds" {
  transfer_id = yandex_datatransfer_transfer.pgpg_transfer.id
}
