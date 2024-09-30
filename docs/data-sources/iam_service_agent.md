---
subcategory: "IAM (Identity and Access Management)"
page_title: "Yandex: yandex_iam_service_agent"
description: |-
  Get information about a Yandex.Cloud Service Agent.
---


# yandex_iam_service_agent




```terraform
data "yandex_iam_user" "admin" {
  login = "my-yandex-login"
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
