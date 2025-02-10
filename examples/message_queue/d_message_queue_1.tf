//
// Get information about existing Message Queue.
//
data "yandex_message_queue" "my_queue" {
  name = "ymq_terraform_example"
}
