//
// Get information about existing IAM Service Agent.
//
data "yandex_iam_service_agent" "my_service_agent" {
  cloud_id        = "some_cloud_id"
  service_id      = "some_service_id"
  microservice_id = "some_microservice_id"
}

output "my_service_agent_id" {
  value = "data.yandex_iam_service_agent.my_service_agent.id"
}
