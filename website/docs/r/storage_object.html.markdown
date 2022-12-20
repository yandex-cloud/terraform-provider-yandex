---
layout: "yandex"
page_title: "Yandex: yandex_storage_object"
sidebar_current: "docs-yandex-storage-object"
description: |-
 Allows management of a Yandex.Cloud Storage Object.
---

# yandex\_storage\_object

Allows management of [Yandex.Cloud Storage Object](https://cloud.yandex.com/docs/storage/concepts/object).

## Example Usage

Example creating an object in an existing `cat-pictures` bucket.

```hcl
resource "yandex_storage_object" "cute-cat-picture" {
  bucket = "cat-pictures"
  key    = "cute-cat"
  source = "/images/cats/cute-cat.jpg"
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the containing bucket.

* `key` - (Required) The name of the object once it is in the bucket.

* `source` - (Optional, conflicts with `content` and `content_base64`) The path to a file that will be read and uploaded as raw bytes for the object content.

* `content` - (Optional, conflicts with `source` and `content_base64`) Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text.

* `content_base64` - (Optional, conflicts with `source` and `content`) Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file.

* `content_type` - (Optional) A standard MIME type describing the format of the object data, e.g. `application/octet-stream`. All Valid MIME Types are valid for this input.

* `access_key` - (Optional) The access key to use when applying changes. If omitted, `storage_access_key` specified in config is used.

* `secret_key` - (Optional) The secret key to use when applying changes. If omitted, `storage_secret_key` specified in config is used.

* `acl` - (Optional) The [predefined ACL](https://cloud.yandex.com/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`.

~> **Note:** To change ACL after creation, the service account to which used access and secret keys correspond should have `storage.admin` role, though this role is not necessary to be able to create an object with any ACL.

* `object_lock_legal_hold_status` - (Optional) Specifies a [legal hold status](https://cloud.yandex.com/en/docs/storage/concepts/object-lock#types) of an object. Requires `object_lock_configuration` to be enabled on a bucket.

* `object_lock_mode` - (Optional) Specifies a type of object lock. One of `["GOVERNANCE", "COMPLIANCE"]`. It must be set simultaneously with `object_lock_retain_until_date`. Requires `object_lock_configuration` to be enabled on a bucket.

* `object_lock_retain_until_date` - (Optional) Specifies date and time in RTC3339 format until which an object is to be locked. It must be set simultaneously with `object_lock_mode`. Requires `object_lock_configuration` to be enabled on a bucket. 

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The `key` of the resource.
