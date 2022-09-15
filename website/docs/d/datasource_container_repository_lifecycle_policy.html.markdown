---
layout: "yandex"
page_title: "Yandex: yandex_container_repository_lifecycle_policy"
sidebar_current: "docs-yandex-datasource-container-repository-lifecycle-policy"
description: |-
  Get information about a Yandex Container Repository Lifecycle Policy.
---

# yandex\_container\_repository\_lifecycle\_policy

Get information about a Yandex Container Repository. For more information, see
[the official documentation](https://cloud.yandex.com/en-ru/docs/container-registry/concepts/lifecycle-policy)

## Example Usage

```hcl
data "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy_by_id" {
  lifecycle_policy_id = yandex_container_repository_lifecycle_policy.my_lifecycle_policy.id
}
```

## Argument Reference

The following arguments are supported:

* `lifecycle_policy_id` - (Optional) The ID of a specific Lifecycle Policy.

* `repository_id` - (Optional) The ID of a repository which Lifecycle Policy belongs to.

* `name` - (Optional) Name of Lifecycle Policy.

~> **NOTE:** Either `lifecycle_policy_id` or `name` and `repository_id` must be specified.

## Attributes Reference

* `description` - Description of the lifecycle policy.

* `status` - The status of lifecycle policy.

created_at

---

The `rule` block supports:

* `description` - Description of the lifecycle policy.

* `expire_period` - The period of time that must pass after creating a image for it to suit the automatic deletion criteria. It must be a multiple of 24 hours.

* `tag_regexp` - Tag to specify a filter as a regular expression. For example `.*` - all images with tags.

* `untagged` - If enabled, rules apply to untagged Docker images.

* `retained_top` - The number of images to be retained even if the expire_period already expired.
