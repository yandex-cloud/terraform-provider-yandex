---
subcategory: "Cloud Domain Name System (DNS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a DNS Recordset within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

Manages a DNS Recordset.

## Example usage

{{ tffile "examples/dns_recordset/r_dns_recordset_1.tf" }}

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required) The id of the zone in which this record set will reside.
* `name` - (Required) The DNS name this record set will apply to.
* `type` - (Required) The DNS record set type.
* `ttl` - (Optional) The time-to-live of this record set (seconds).
* `data` - (Optional) The string data for the records in this record set.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/dns_recordset/import.sh" }}
