---
layout: "yandex"
page_title: "Yandex: yandex_datasphere_project_iam_binding"
sidebar_current: "docs-yandex-datasphere-project-iam-binding"
description: |-
  Allows management of a single IAM binding for a Yandex Datasphere Project.
---

## yandex\_datasphere\_project\_iam\_binding

```hcl
resource "yandex_datasphere_project_iam_binding" "project-iam" {
  project_id = "your-datasphere-project-id"
  role        = "datasphere.community-projects.developer"
  members = [
    "system:allUsers",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required) The Yandex Cloud Datasphere Project ID to apply a binding to.

* `role` - (Required) The role that should be applied. See [roles](https://cloud.yandex.com/en/docs/datasphere/security/)

* `members` - (Required) Identities that will be granted the privilege in `role`.
  Each entry can have one of the following values:
    * **userAccount:{user_id}**: A unique user ID that represents a specific Yandex account.
    * **serviceAccount:{service_account_id}**: A unique service account ID.
    * **federatedUser:{federated_user_id}:**: A unique saml federation user account ID.
    * **group:{group_id}**: A unique group ID.
    * **system:group:federation:{federation_id}:users**: All users in federation.
    * **system:group:organization:{organization_id}:users**: All users in organization.
    * **system:allAuthenticatedUsers**: All authenticated users. 
    * **system:allUsers**: All users, including unauthenticated ones.

    Note: for more information about system groups, see the [documentation](https://cloud.yandex.com/docs/iam/concepts/access-control/system-group).
