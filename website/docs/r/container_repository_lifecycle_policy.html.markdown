---
layout: "yandex"
page_title: "Yandex: yandex_container_repository_lifecycle_policy"
sidebar_current: "docs-yandex-container-repository-lifecycle-policy"
description: |-
  Creates a new container repository lifecycle policy.
---

# yandex\_container\_repository\_lifecycle\_policy

Creates a new container repository lifecycle policy. For more information, see
[the official documentation](https://cloud.yandex.com/en-ru/docs/container-registry/concepts/lifecycle-policy)

## Example Usage

```hcl
resource "yandex_container_registry" "my_registry" {
  name = "test-registry"
}

resource "yandex_container_repository" "my_repository" {
  name = "${yandex_container_registry.my_registry.id}/test-repository"
}

resource "yandex_container_repository_lifecycle_policy" "my_lifecycle_policy" {
  name          = "test-lifecycle-policy-name"
  status        = "active"
  repository_id = yandex_container_repository.my_repository.id

  rule {
    description  = "my description"
    untagged     = true
    tag_regexp   = ".*"
    retained_top = 1
  }
}
```

## Argument Reference

The following arguments are supported:

* `repository_id` - (Required) The ID of the repository that the resource belongs to.

* `status` - (Required) The status of lifecycle policy. Must be `active` or `disabled`.

* `name` - (Optional) Lifecycle policy name.

* `description` - (Optional) Description of the lifecycle policy.

---

The `rule` block supports:

* `description` - (Optional) Description of the lifecycle policy.

* `expire_period` - (Optional) The period of time that must pass after creating a image for it to suit the automatic deletion criteria. It must be a multiple of 24 hours.

* `tag_regexp` - (Optional) Tag to specify a filter as a regular expression. For example `.*` - all images with tags.

* `untagged` - (Optional) If enabled, rules apply to untagged Docker images.

* `retained_top` - (Optional) The number of images to be retained even if the expire_period already expired.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `created_at` - Creation timestamp of the instance.

## Import

A lifecycle policy can be imported using the `id` of the resource, e.g.

```bash
terraform import yandex_container_repository_lifecycle_policy.my_lifecycle_policy lifecycle_policy_id
```
