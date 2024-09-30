---
subcategory: "DNS (Domain Name System)"
page_title: "Yandex: yandex_dns_recordset"
description: |-
  Manages a DNS Recordset within Yandex.Cloud.
---


# yandex_dns_recordset




Manages a DNS Recordset.

```terraform
resource "yandex_dns_zone" "zone1" {
  name = "my-private-zone"
  zone = "example.com."
}

resource "yandex_dns_zone_iam_binding" "viewer" {
  dns_zone_id = yandex_dns_zone.zone1.id
  role        = "dns.viewer"
  members     = ["userAccount:foo_user_id"]
}
```

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required) The id of the zone in which this record set will reside.
* `name` - (Required) The DNS name this record set will apply to.
* `type` - (Required) The DNS record set type.
* `ttl` - (Optional) The time-to-live of this record set (seconds).
* `data` - (Optional) The string data for the records in this record set.

## Import

DNS recordset can be imported using this format:

```
$ terraform import yandex_dns_recordset.rs1 {zone_id}/{name}/{type}
```
