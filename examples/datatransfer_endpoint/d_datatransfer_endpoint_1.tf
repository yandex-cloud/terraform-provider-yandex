//
// Get information about existing Datatransfer Endpoint
//
data "yandex_datatransfer_endpoint" "pg_source_ds" {
  endpoint_id = yandex_datatransfer_endpoint.pg_source.id
}
