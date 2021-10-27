---
layout: "yandex"
page_title: "Yandex: yandex_vpc_security_group_rule"
sidebar_current: "docs-yandex-vpc-security-group-rule"
description: |-
  Yandex VPC Security Group Rule.
---

# yandex\_vpc\_security\_group\_rule

Manages a single Secuirity Group Rule within the Yandex.Cloud. For more information, see the official documentation
of [security groups](https://cloud.yandex.com/docs/vpc/concepts/security-groups)
and [security group rules](https://cloud.yandex.com/docs/vpc/concepts/security-groups#rules).

~> **NOTE:** There is another way to manage security group rules by `ingress` and `egress` arguments in [yandex_vpc_security_group](vpc_security_group.html). Both ways are equivalent but not compatible now. Using in-line rules of [yandex_vpc_security_group](vpc_security_group.html) with Security Group Rule resource at the same time will cause a conflict of rules configuration.

## Example Usage

```hcl
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_security_group" "group1" {
  name        = "My security group"
  description = "description for my security group"
  network_id  = "${yandex_vpc_network.lab-net.id}"

  labels = {
    my-label = "my-label-value"
  }
}

resource "yandex_vpc_security_group_rule" "rule1" {
  security_group_binding = yandex_vpc_security_group.group1.id
  direction              = "ingress"
  description            = "rule1 description"
  v4_cidr_blocks         = ["10.0.1.0/24", "10.0.2.0/24"]
  port                   = 8080
  protocol               = "TCP"
}

resource "yandex_vpc_security_group_rule" "rule2" {
  security_group_binding = yandex_vpc_security_group.group1.id
  direction              = "egress"
  description            = "rule2 description"
  v4_cidr_blocks         = ["10.0.1.0/24"]
  from_port              = 8090
  to_port                = 8099
  protocol               = "UDP"
}
```

## Argument Reference

The following arguments are supported:

* `security_group_binding` (Required) - ID of the security group this rule belongs to.
* `direction` (Required) - direction of the rule. Can be `ingress` (inbound) or `egress` (outbound).
* `protocol` (Required) - One of `ANY`, `TCP`, `UDP`, `ICMP`, `IPV6_ICMP`.

* `description` (Optional) - Description of the rule.
* `labels` (Optional) - Labels to assign to this rule.
* `from_port` (Optional) - Minimum port number.
* `to_port` (Optional) - Maximum port number.
* `port` (Optional) - Port number (if applied to a single port).
* `security_group_id` (Optional) - Target security group ID for this rule.
* `predefined_target` (Optional) - Special-purpose targets such as "self_security_group". [See docs](https://cloud.yandex.com/docs/vpc/concepts/security-groups) for possible options.
* `v4_cidr_blocks` (Optional) - The blocks of IPv4 addresses for this rule.
* `v6_cidr_blocks` (Optional) - The blocks of IPv6 addresses for this rule. `v6_cidr_blocks` argument is currently not supported. It will be available in the future.

~> **NOTE:** Either one `port` argument or both `from_port` and `to_port` arguments can be specified.


~> **NOTE:** If `port` or `from_port`/`to_port` aren't specified or set by -1, ANY port will be sent.


~> **NOTE:** Can't use specified port if protocol is one of `ICMP` or `IPV6_ICMP`.


## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Id of the rule.
