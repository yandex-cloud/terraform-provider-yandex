---
layout: "yandex"
page_title: "Yandex: yandex_vpc_security_group"
sidebar_current: "docs-yandex-vpc-security-group"
description: |-
  Yandex VPC Security Group.
---

# yandex\_vpc\_security\_group

Manages a Security Group within the Yandex.Cloud. For more information, see
[the official documentation](https://cloud.yandex.com/docs/vpc/concepts).

Security groups is in private preview phase and not available right now. Please wait for public preview announcement.


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

  rule {
    direction      = "INGRESS"
    protocol       = "TCP"
    description    = "rule1 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  rule {
    direction      = "EGRESS"
    protocol       = "ANY"
    description    = "rule2 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }

  rule {
    direction      = "EGRESS"
    protocol       = "27"
    description    = "rule3 description"
    v4_cidr_blocks = ["10.0.1.0/24"]
    from_port      = 8090
    to_port        = 8099
  }
}
```

## Argument Reference

The following arguments are supported:

* `network_id` (Required) - ID of the network this security group belongs to.

---

* `name` (Optional) - Name of the security group.
* `description` (Optional) - Description of the security group.
* `folder_id` (Optional) - ID of the folder this security group belongs to.
* `labels` (Optional) - Labels to assign to this security group.
* `rule` (Optional) - A list of rules.

The structure is documented below.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - Status of this security group.
* `created_at` - Creation timestamp of this security group.

---

The `rule` block supports:

* `direction` (Required) - Direction of the rule. One of `INGRESS` or `EGRESS` should be specified.

---

* `protocol` (Required) - One of `ANY`, `TCP`, `UDP`, `ICMP`, `IPV6_ICMP` or protocol number..
* `description` (Optional) - Description of the rule.
* `labels` (Optional) - Labels to assign to this rule.
* `protocol_number` (Optional) - Number of the protocol defined by [IANA](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml). Values are `0`,`6`,`17`.
* `from_port` (Optional) - Minimum port number.
* `to_port` (Optional) - Maximum port number.
* `port` (Optional) - Port number (if applied to a single port).
* `v4_cidr_blocks` (Optional) - The blocks of IPv4 addresses for this rule.
* `v6_cidr_blocks` (Optional) - The blocks of IPv6 addresses for this rule. `v6_cidr_blocks` argument is currently not supported. It will be available in the future.


~> **NOTE:** Only one of `protocol_name` or `protocol_number` can be specified. If none of them is set, all protocols are allowed.
~> **NOTE:** Either one `port` argument or both `from_port` and `to_port` arguments can be specified.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Id of the rule.
