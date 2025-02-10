//
// Get information about existing NLB Target Group.
//
data "yandex_lb_target_group" "my_tg" {
  target_group_id = "my-target-group-id"
}
