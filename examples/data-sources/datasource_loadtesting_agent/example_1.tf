data "yandex_loadtesting_agent" "my_agent" {
  agent_id = "some_agent_id"
}

output "instance_external_ip" {
  value = data.yandex_loadtesting_agent.my_agent.compute_instance.0.network_interface.0.nat_ip_address
}
