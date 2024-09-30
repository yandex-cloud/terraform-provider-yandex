resource "yandex_message_queue" "example_fifo_queue" {
  name                        = "ymq_terraform_fifo_example.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
