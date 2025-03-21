---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about a Yandex Certificate Manager Certificate.
---

# {{.Name}} ({{.Type}})

Get information about a Yandex Certificate Manager Certificate. For more information, see [the official documentation](https://yandex.cloud/docs/certificate-manager/concepts/).

## Example usage

{{ tffile "examples/cm_certificate/d_cm_certificate_1.tf" }}

{{ tffile "examples/cm_certificate/d_cm_certificate_2.tf" }}

## Argument Reference

The following arguments are supported:

* `certificate_id` (Optional) - Certificate Id.
* `name` - (Optional) - Name of the Certificate.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `wait_validation` - (Optional, default is `false`) If `true`, the operation won't be completed while the certificate is in `VALIDATING`.

~> One of `certificate_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - Certificate Id.
* `description` - Certificate description.
* `labels` - Labels to assign to this certificate.
* `created_at` - Certificate create timestamp.
* `updated_at` - Certificate update timestamp.
* `type` - Certificate type: `"MANAGED"` or `"IMPORTED"`.
* `status` - Certificate status: `"VALIDATING"`, `"INVALID"`, `"ISSUED"`, `"REVOKED"`, `"RENEWING"` or `"RENEWAL_FAILED"`.
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
