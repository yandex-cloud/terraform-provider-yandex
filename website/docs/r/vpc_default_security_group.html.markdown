---
layout: "yandex"
page_title: "Yandex: yandex_vpc_default_security_group"
sidebar_current: "docs-yandex-vpc-default-security-group"
description: |-
  Yandex VPC Default Security Group.
---

# yandex\_vpc\_default\_security\_group

Manages a Default Security Group within the Yandex.Cloud. For more information, see the official documentation 
of [security group](https://cloud.yandex.com/docs/vpc/concepts/security-groups) 
or [default security group](https://cloud.yandex.com/docs/vpc/concepts/security-groups#default-security-group).

~> **NOTE:** This resource is not intended for managing security group in general case. To manage normal security group use [yandex_vpc_security_group](vpc_security_group.html)

When [network](https://cloud.yandex.com/docs/vpc/concepts/network) is created, a non-removable security group, called a *default security group*, is automatically attached to it.
Life time of default security group cannot be controlled, so in fact the resource `yandex_vpc_default_security_group` does not create or delete any security groups, instead it simply takes or releases
control of the default security group.

~> **NOTE:** When Terraform takes over management of the default security group, it **deletes** all info in it (including security group rules) and replace it with specified configuration. When Terraform drops the management (i.e. when resource is deleted from statefile and management), the state of the security group **remains the same** as it was before the deletion.

~> **NOTE:** Duplicating a resource (specifying same `network_id` for two different default security groups) will cause errors in the apply stage of your's configuration.

## Example Usage

```hcl
resource "yandex_vpc_network" "lab-net" {
  name = "lab-network"
}

resource "yandex_vpc_default_security_group" "default-sg" {
  description = "description for default security group"
  network_id  = "${yandex_vpc_network.lab-net.id}"

  labels = {
    my-label = "my-label-value"
  }

  ingress {
    protocol       = "TCP"
    description    = "rule1 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    port           = 8080
  }

  egress {
    protocol       = "ANY"
    description    = "rule2 description"
    v4_cidr_blocks = ["10.0.1.0/24", "10.0.2.0/24"]
    from_port      = 8090
    to_port        = 8099
  }

  egress {
    protocol       = "UDP"
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

* `description` (Optional) - Description of the security group.
* `folder_id` (Optional) - ID of the folder this security group belongs to.
* `labels` (Optional) - Labels to assign to this security group.
* `ingress` (Optional) - A list of ingress rules.
* `egress` (Optional) - A list of egress rules. The structure is documented below.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `name` - Name of this security group.
* `status` - Status of this security group.
* `created_at` - Creation timestamp of this security group.

---

The `ingress` and `egress` block supports:

* `protocol` (Required) - One of `ANY`, `TCP`, `UDP`, `ICMP`, `IPV6_ICMP`.
* `description` (Optional) - Description of the rule.
* `labels` (Optional) - Labels to assign to this rule.
* `from_port` (Optional) - Minimum port number.
* `to_port` (Optional) - Maximum port number.
* `port` (Optional) - Port number (if applied to a single port).
* `security_group_id` (Optional) - Target security group ID for this rule.
* `predefined_target` (Optional) - Special-purpose targets. `self_security_group` refers to this particular security group. `loadbalancer_healthchecks` represents [loadbalancer health check nodes](https://cloud.yandex.com/docs/network-load-balancer/concepts/health-check).
* `v4_cidr_blocks` (Optional) - The blocks of IPv4 addresses for this rule.
* `v6_cidr_blocks` (Optional) - The blocks of IPv6 addresses for this rule. `v6_cidr_blocks` argument is currently not supported. It will be available in the future.


~> **NOTE:** Either one `port` argument or both `from_port` and `to_port` arguments can be specified.
~> **NOTE:** If `port` or `from_port`/`to_port` aren't specified or set by -1, ANY port will be sent.
~> **NOTE:** Can't use specified port if protocol is one of `ICMP` or `IPV6_ICMP`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Id of the security group.
