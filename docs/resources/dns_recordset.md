---
subcategory: "Cloud Domain Name System (DNS)"
page_title: "Yandex: yandex_dns_recordset"
description: |-
  Manages a DNS Recordset within Yandex Cloud.
---

# yandex_dns_recordset (Resource)

Manages a DNS Recordset.

## Example usage

```terraform
//
// Create a new DNS Zone & two DNS Records on it.
//
resource "yandex_dns_zone" "zone1" {
  name        = "my_private_zone"
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

resource "yandex_dns_recordset" "rs2" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv2"
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.2"]
}

// Auxiliary resource for DNS Zone
resource "yandex_vpc_network" "foo" {}
```

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required) The id of the zone in which this record set will reside.
* `name` - (Required) The DNS name this record set will apply to.
* `type` - (Required) The DNS record set type.
* `ttl` - (Optional) The time-to-live of this record set (seconds).
* `data` - (Optional) The string data for the records in this record set.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_dns_recordset.<resource Name> <{zone_id}/{name}/{type}>
terraform import yandex_dns_recordset.rs1 dns9m**********tducf/srv.example.com./A
```
