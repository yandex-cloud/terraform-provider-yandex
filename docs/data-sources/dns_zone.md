---
subcategory: "Cloud Domain Name System (DNS)"
page_title: "Yandex: yandex_dns_zone"
description: |-
  Get information about a DNS Zone within Yandex Cloud.
---

# yandex_dns_zone (Data Source)

Get information about a DNS Zone.

## Example usage

```terraform
//
// Get information about existing DNS Zone.
//
data "yandex_dns_zone" "foo" {
  dns_zone_id = yandex_dns_zone.zone1.id
}

output "zone" {
  value = data.yandex_dns_zone.foo.zone
}
```

## Argument Reference

* `dns_zone_id` - (Optional) The ID of the DNS Zone.
* `name` - (Optional) - Name of the DNS Zone.

~> One of `dns_zone_id` or `name` should be specified.

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

## Attributes Reference

* `zone` - (Computed) The DNS name of this zone, e.g. "example.com.". Must ends with dot.
* `folder_id` - (Computed) The ID of the folder that the resource belongs to. If it is not provided, the default provider folder is used.
* `name` - (Computed) User assigned name of a specific resource. Must be unique within the folder.
* `description` - (Computed) Description of the DNS zone.
* `labels` - (Computed) A set of key/value label pairs to assign to the DNS zone.
* `public` - (Computed) The zone's visibility: public zones are exposed to the Internet, while private zones are visible only to Virtual Private Cloud resources.
* `private_networks` - (Computed) For privately visible zones, the set of Virtual Private Cloud resources that the zone is visible from.
* `created_at` - (Computed) The DNS zone creation timestamp.
* `deletion_protection` - (Computed) Flag that protects the dns zone from accidental deletion.
