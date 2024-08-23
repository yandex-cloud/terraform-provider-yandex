---
layout: "yandex"
page_title: "Yandex: yandex_iam_service_agent"
sidebar_current: "docs-yandex-datasource-iam-service-agent"
description: |-
  Get information about a Yandex.Cloud Service Agent.
---

# yandex\_iam\_service\_agent

## Example Usage

```hcl
data "yandex_iam_service_agent" my_service_agent {
  cloud_id        = "some_cloud_id"
  sevice_id       = "som_service_id"
  microservice_id = "some_microservice_id"
}

output "my_service_agent_id" {
  value = "data.yandex_iam_service_agent.my_service_agent.id"
}
```

## Argument Reference

* `cloud_id` - (Required) ID of the cloud.
* `service_id` - (Required) ID of the service-control service.
* `microservice_id` - (Required) ID of the service-control microservice.

## Attributes Reference

The following attributes are exported:

* `service_account_id` - ID of the resolved agent service account.
* `service_id` - ID of the service-control service.
* `microservice_id` - ID of the service-control microservice.
