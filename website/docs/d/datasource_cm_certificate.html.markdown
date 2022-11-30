---
layout: "yandex"
page_title: "Yandex: yandex_cm_certificate"
sidebar_current: "docs-yandex-datasource-cm-certificate"
description: |-
Get information about a Yandex Certificate Manager Certificate.
---

# yandex\_cm\_certificate

Get information about a Yandex Certificate Manager Certificate.
For more information, see [the official documentation](https://cloud.yandex.com/en/docs/certificate-manager/concepts/).

```hcl
data "yandex_cm_certificate" "example_by_id" {
  certificate_id = "certificate-id"
}

data "yandex_cm_certificate" "example_by_name" {
  folder_id = "folder-id"
  name      = "example"
}
```

This data source is used to define [Certificate Manager Certificate](https://cloud.yandex.com/en/docs/certificate-manager/concepts/) that can be used by other resources.
Can also be used to wait for certificate validation.

## Example Usage of Certificate Validation Wait

```hcl
resource "yandex_cm_certificate" "example" {
  name    = "example"
  domains = ["example.com", "*.example.com"]

  managed {
    challenge_type  = "DNS_CNAME"
    challenge_count = 1 # "example.com" and "*.example.com" has the same challenge
  }
}

resource "yandex_dns_recordset" "example" {
  count   = yandex_cm_certificate.example.managed[0].challenge_count
  zone_id = "example-zone-id"
  name    = yandex_cm_certificate.example.challenges[count.index].dns_name
  type    = yandex_cm_certificate.example.challenges[count.index].dns_type
  data    = [yandex_cm_certificate.example.challenges[count.index].dns_value]
  ttl     = 60
}

data "yandex_cm_certificate" "example" {
  depends_on      = [yandex_dns_recordset.example]
  certificate_id  = yandex_cm_certificate.example.id
  wait_validation = true
}

# Use data.yandex_cm_certificate.example.id to get validated certificate
```

## Argument Reference

The following arguments are supported:

* `certificate_id` (Optional) - Certificate Id.
* `name` - (Optional) - Name of the Certificate.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `wait_validation` - (Optional, default is `false`) If `true`, the operation won't be completed while the certificate is in `VALIDATING`.

~> **NOTE:** One of `certificate_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - Certificate Id.
* `description` - Certificate description.
* `labels` - Labels to assign to this certificate.
* `created_at` - Certificate create timestamp.
* `updated_at` - Certificate update timestamp.
* `type` - Certificate type: `"MANAGED"` or `"IMPORTED"`.
* `status` - Certificate status: `"VALIDATING"`, `"INVALID"`,  `"ISSUED"`, `"REVOKED"`, `"RENEWING"` or `"RENEWAL_FAILED"`.
* `issuer` - Certificate issuer.
* `subject` - Certificate subject.
* `serial` - Certificate serial number.
* `issued_at` - Certificate issue timestamp.
* `not_before` - Certificate start valid period.
* `not_after` - Certificate end valid period.
* `challenges` - Array of challenges. Structure is documented below.

The `challenges` block represents (for each array element):

* `domain` - Validated domain.
* `type` - Challenge type `"DNS"` or `"HTTP"`.
* `created_at` - Time the challenge was created.
* `updated_at` - Last time the challenge was updated.
* `message` - Current status message.
* `dns_name` - DNS record name (only for DNS challenge).
* `dns_type` - DNS record type: `"TXT"` or `"CNAME"` (only for DNS challenge).
* `dns_value` - DNS record value (only for DNS challenge).
* `http_url` - URL where the challenge content `http_content` should be placed (only for HTTP challenge).
* `http_content` - The content that should be made accessible with the given `http_url` (only for HTTP challenge).
