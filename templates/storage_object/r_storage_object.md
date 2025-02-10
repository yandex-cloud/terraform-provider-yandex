---
subcategory: "Object Storage (S3)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud Storage Object.
---

# {{.Name}} ({{.Type}})

Allows management of [Yandex Cloud Storage Object](https://yandex.cloud/docs/storage/concepts/object).

## Example usage

{{ tffile "examples/storage_object/r_storage_object_1.tf" }}

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the containing bucket.

* `key` - (Required) The name of the object once it is in the bucket.

* `source` - (Optional, conflicts with `content` and `content_base64`) The path to a file that will be read and uploaded as raw bytes for the object content.

* `source_hash` - (Optional) Used to trigger object update when the source content changes. So the only meaningful value is `filemd5("path/to/source")` (The value is only stored in state and not saved by Yandex Storage).

* `content` - (Optional, conflicts with `source` and `content_base64`) Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text.

* `content_base64` - (Optional, conflicts with `source` and `content`) Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file.

* `content_type` - (Optional) A standard MIME type describing the format of the object data, e.g. `application/octet-stream`. All Valid MIME Types are valid for this input.

* `access_key` - (Optional) The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.

* `secret_key` - (Optional) The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.

* `acl` - (Optional) The [predefined ACL](https://yandex.cloud/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`.

~> To change ACL after creation, the service account to which used access and secret keys correspond should have `storage.admin` role, though this role is not necessary to be able to create an object with any ACL.

* `object_lock_legal_hold_status` - (Optional) Specifies a [legal hold status](https://yandex.cloud/docs/storage/concepts/object-lock#types) of an object. Requires `object_lock_configuration` to be enabled on a bucket.

* `object_lock_mode` - (Optional) Specifies a type of object lock. One of `["GOVERNANCE", "COMPLIANCE"]`. It must be set simultaneously with `object_lock_retain_until_date`. Requires `object_lock_configuration` to be enabled on a bucket.

* `object_lock_retain_until_date` - (Optional) Specifies date and time in RTC3339 format until which an object is to be locked. It must be set simultaneously with `object_lock_mode`. Requires `object_lock_configuration` to be enabled on a bucket.

* `tags` - (Optional) Specifies an object tags.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The `key` of the resource.

## Import

~> Import for this resource is not implemented yet.

