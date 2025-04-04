---
subcategory: "Object Storage (S3)"
page_title: "Yandex: yandex_storage_object"
description: |-
  Allows management of a Yandex Cloud Storage Object.
---

# yandex_storage_object (Resource)

Allows management of [Yandex Cloud Storage Object](https://yandex.cloud/docs/storage/concepts/object).

## Example usage

```terraform
//
// Create a new Storage Object in Bucket.
//
resource "yandex_storage_object" "cute-cat-picture" {
  bucket = "cat-pictures"
  key    = "cute-cat"
  source = "/images/cats/cute-cat.jpg"
  tags = {
    test = "value"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `bucket` (String) The name of the containing bucket.
- `key` (String) The name of the object once it is in the bucket.

### Optional

- `access_key` (String) The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.
- `acl` (String) The [predefined ACL](https://yandex.cloud/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`.

~> To change ACL after creation, the service account to which used access and secret keys correspond should have `storage.admin` role, though this role is not necessary to be able to create an object with any ACL.
- `content` (String) Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text. Conflicts with `source` and `content_base64`.
- `content_base64` (String) Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file. Conflicts with `source` and `content`.
- `content_type` (String) A standard MIME type describing the format of the object data, e.g. `application/octet-stream`. All Valid MIME Types are valid for this input.
- `object_lock_legal_hold_status` (String) Specifies a [legal hold status](https://yandex.cloud/docs/storage/concepts/object-lock#types) of an object. Requires `object_lock_configuration` to be enabled on a bucket.
- `object_lock_mode` (String) Specifies a type of object lock. One of `["GOVERNANCE", "COMPLIANCE"]`. It must be set simultaneously with `object_lock_retain_until_date`. Requires `object_lock_configuration` to be enabled on a bucket.
- `object_lock_retain_until_date` (String) Specifies date and time in RTC3339 format until which an object is to be locked. It must be set simultaneously with `object_lock_mode`. Requires `object_lock_configuration` to be enabled on a bucket.
- `secret_key` (String, Sensitive) The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.
- `source` (String) The path to a file that will be read and uploaded as raw bytes for the object content. Conflicts with `content` and `content_base64`.
- `source_hash` (String) Used to trigger object update when the source content changes. So the only meaningful value is `filemd5("path/to/source"). The value is only stored in state and not saved by Yandex Storage.
- `tags` (Map of String) The `tags` object for setting tags (or labels) for bucket. See [Tags](https://yandex.cloud/docs/storage/concepts/tags) for more information.

### Read-Only

- `id` (String) The ID of this resource.

## Import

~> Import for this resource is not implemented yet.