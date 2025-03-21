---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  Get content from a Yandex Certificate Manager Certificate.
---

# {{.Name}} ({{.Type}})

Get content (certificate, private key) from a Yandex Certificate Manager Certificate. For more information, see [the official documentation](https://yandex.cloud/docs/certificate-manager/concepts/).

## Example usage

{{ tffile "examples/cm_certificate_content/d_cm_certificate_content_1.tf" }}

## Argument Reference

The following arguments are supported:

* `certificate_id` (Optional) - Certificate Id.
* `name` - (Optional) - Certificate name.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `wait_validation` - (Optional, default is `false`) If `true`, the operation won't be completed while the certificate is in `VALIDATING`.
* `private_key_format` - (Optional) Format in which you want to export the private_key: `"PKCS1"` or `"PKCS8"`.

~> One of `certificate_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - Certificate Id.
* `certificates` - List of certificates in chain.
* `private_key` - Private key in specified format.
