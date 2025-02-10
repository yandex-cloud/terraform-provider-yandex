//
// Get information about existing ALB Target Group
//
data "yandex_alb_target_group" "foo" {
  target_group_id = "my-target-group-id"
}
