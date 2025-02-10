//
// Get information about existing Network Load Balancer (NLB).
//
data "yandex_lb_network_load_balancer" "my_nlb" {
  network_load_balancer_id = "my-network-load-balancer"
}
