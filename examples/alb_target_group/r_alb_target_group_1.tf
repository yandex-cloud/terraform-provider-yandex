resource "yandex_alb_target_group" "foo" {
  name = "my-target-group"

  target {
    subnet_id  = yandex_vpc_subnet.my-subnet.id
    ip_address = yandex_compute_instance.my-instance-1.network_interface.0.ip_address
  }

  target {
    subnet_id  = yandex_vpc_subnet.my-subnet.id
    ip_address = yandex_compute_instance.my-instance-2.network_interface.0.ip_address
  }
}
