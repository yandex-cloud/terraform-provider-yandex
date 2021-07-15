---
layout: "yandex"
page_title: "Yandex: yandex_dns_zone"
sidebar_current: "docs-yandex-dns-zone"
description: |-
Manages a DNS Zone within Yandex.Cloud.
---

# yandex\_dns\_zone

Manages a DNS Zone.

## Example Usage

```hcl
resource "yandex_vpc_network" "foo" {}

resource "yandex_dns_zone" "zone1" {
  name        = "my-private-zone"
  description = "desc"

  labels = {
    label1 = "label-1-value"
  }

  zone             = "example.com."
  public           = false
  private_networks = [yandex_vpc_network.foo.id]
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.example.com."
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.1"]
}
```

## Argument Reference

The following arguments are supported:

* `zone` - (Required) The DNS name of this zone, e.g. "example.com.". Must ends with dot.
* `folder_id` - (Optional) ID of the folder to create a zone in. If it is not provided, the default provider folder is used.
* `name` - (Optional) User assigned name of a specific resource. Must be unique within the folder.
* `description` - (Optional) Description of the DNS zone.
* `labels` - (Optional) A set of key/value label pairs to assign to the DNS zone.
* `public` - (Optional) The zone's visibility: public zones are exposed to the Internet, while private zones are visible only to Virtual Private Cloud resources.
* `private_networks` - (Optional) For privately visible zones, the set of Virtual Private Cloud resources that the zone is visible from.

## Attributes Reference

* `id` - (Computed) ID of a new DNS zone.
* `created_at` - (Computed) The DNS zone creation timestamp.
