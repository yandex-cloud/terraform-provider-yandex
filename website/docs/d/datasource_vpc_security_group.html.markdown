---
layout: "yandex"
page_title: "Yandex: yandex_vpc_security_group"
sidebar_current: "docs-yandex-datasource-vpc-security-group"
description: |-
  Get information about a Yandex VPC Security Group.
---

# yandex\_vpc\_security\_group

Get information about a Yandex VPC Security Group. For more information, see
[Yandex.Cloud VPC](https://cloud.yandex.com/docs/vpc/concepts/index).

```hcl
data "yandex_vpc_security_group" "group1" {
  security_group_id = "my-id"
}
```

This data source is used to define Security Group that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `security_group_id` (Required) - Security Group ID.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

The following attribute is exported:

* `name` - Name of the security group.
* `description` - Description of the security group.
* `network_id` - ID of the network this security group belongs to.
* `folder_id` - ID of the folder this security group belongs to.
* `labels` - Labels to assign to this security group.
* `ingress` - A list of ingress rules. The structure is documented below.
* `egress` - A list of egress rules. The structure is documented below.
* `status` - Status of this security group.
* `created_at` - Creation timestamp of this security group.

---

The `ingress` and `egress` block supports:
* `id` - Id of the rule.
* `description` - Description of the rule.
* `labels` - Labels to assign to this rule.
* `protocol` - One of `ANY`, `TCP`, `UDP`, `ICMP`, `IPV6_ICMP`.
* `from_port` - Minimum port number.
* `to_port` - Maximum port number.
* `port` - Port number (if applied to a single port).
* `security_group_id` - Target security group ID for this rule.
* `predefined_target` - Specially target for this rule. See docs for choices.
* `v4_cidr_blocks` - The blocks of  IPv4 addresses for this rule.
* `v6_cidr_blocks` - The blocks of  IPv6 addresses for this rule.
