---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: yandex_iam_workload_identity_federated_credential"
description: |-
  Get information about a Yandex IAM federated credential.
---

# yandex_iam_workload_identity_federated_credential (Data Source)

Get information about a [Yandex Cloud IAM federated credential](https://yandex.cloud/docs/iam/concepts/workload-identity#federated-credentials).

## Example Usage

```terraform
data "yandex_iam_workload_identity_federated_credential" "fc" {
  federated_credential_id = "some_fed_cred_id"
}
```
