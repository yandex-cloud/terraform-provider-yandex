//
// Get information about existing Connection.
//
data "yandex_connectionmanager_connection" "my_connection" {
  connection_id = "some_connection_id"
}