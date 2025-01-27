---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Cloud Service Agent.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Cloud Service Agent.

## Example usage

{{ tffile "examples/iam_service_agent/d_iam_service_agent_1.tf" }}

## Argument Reference

* `cloud_id` - (Required) ID of the cloud.
* `service_id` - (Required) ID of the service-control service.
* `microservice_id` - (Required) ID of the service-control microservice.

## Attributes Reference

The following attributes are exported:

* `service_account_id` - ID of the resolved agent service account.
* `service_id` - ID of the service-control service.
* `microservice_id` - ID of the service-control microservice.
