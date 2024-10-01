resource "yandex_message_queue" "example_queue" {
  name                       = "ymq_terraform_example"
  visibility_timeout_seconds = 600
  receive_wait_time_seconds  = 20
  message_retention_seconds  = 1209600
  redrive_policy = jsonencode({
    deadLetterTargetArn = yandex_message_queue.example_deadletter_queue.arn
    maxReceiveCount     = 3
  })
}

resource "yandex_message_queue" "example_deadletter_queue" {
  name = "ymq_terraform_deadletter_example"
}
