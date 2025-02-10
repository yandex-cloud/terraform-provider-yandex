//
// Create a new Cloud Function Trigger.
//
resource "yandex_function_trigger" "my_trigger" {
  name        = "some_name"
  description = "any description"
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id = "tf-test"
  }
}
