data "yandex_compute_instance_group" "my_group" {
  instance_group_id = "some_instance_group_id"
}

output "instance_external_ip" {
  value = data.yandex_compute_instance_group.my_group.instances.*.network_interface.0.nat_ip_address
}
