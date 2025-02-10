---
subcategory: "Certificate Manager"
page_title: "Yandex: yandex_cm_certificate_content"
description: |-
  Get content from a Yandex Certificate Manager Certificate.
---

# yandex_cm_certificate_content (Data Source)

Get content (certificate, private key) from a Yandex Certificate Manager Certificate. For more information, see [the official documentation](https://yandex.cloud/docs/certificate-manager/concepts/).

## Example usage

```terraform
// 
// Get CM Certificate payload. Can be used for Certificate Validation.
//
data "yandex_cm_certificate_content" "example_by_id" {
  certificate_id = "certificate-id"
}

data "yandex_cm_certificate_content" "example_by_name" {
  folder_id = "folder-id"
  name      = "example"
}
```

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
