---
layout: "yandex"
page_title: "Yandex: yandex_storage_bucket"
sidebar_current: "docs-yandex-storage-bucket"
description: |-
 Allows management of a Yandex.Cloud Storage Bucket.
---

# yandex\_storage\_bucket

Allows management of [Yandex.Cloud Storage Bucket](https://cloud.yandex.com/docs/storage/concepts/bucket).

## Example Usage

### Simple Private Bucket

```hcl
resource "yandex_storage_bucket" "test" {
  bucket = "tf-test-bucket"
}
```

### Static Website Hosting

```hcl
resource "yandex_storage_bucket" "test" {
  bucket = "storage-website-test.hashicorp.com"
  acl    = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://storage-website-test.hashicorp.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

### Using ACL policy grants

```hcl
resource "yandex_storage_bucket" "test" {
  bucket = "mybucket"

  grant {
    id          = "myuser"
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL"]
  }

  grant {
    type        = "Group"
    permissions = ["READ", "WRITE"]
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Optional, Forces new resource) The name of the bucket. If omitted, Terraform will assign a random, unique name.

* `bucket_prefix` - (Optional, Forces new resource) Creates a unique bucket name beginning with the specified prefix. Conflicts with `bucket`.

* `access_key` - (Optional) The access key to use when applying changes. If omitted, `storage_access_key` specified in provider config is used.

* `secret_key` - (Optional) The secret key to use when applying changes. If omitted, `storage_secret_key` specified in provider config is used.

* `acl` - (Optional) The [predefined ACL](https://cloud.yandex.com/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`. Conflicts with `grant`.

~> **Note:** To change ACL after creation, the service account to which used access and secret keys correspond should have `admin` role, though this role is not necessary to be able to create a bucket with any ACL.

* `grant` - (Optional) An [ACL policy grant](https://cloud.yandex.ru/docs/storage/concepts/acl#permissions-types). Conflicts with `acl`.

* `force_destroy` - (Optional, Default: `false`) A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable.

* `website` - (Optional) A website object (documented below).

* `cors_rule` - (Optional) A rule of [Cross-Origin Resource Sharing](https://cloud.yandex.com/docs/storage/cors/) (documented below).

The `website` object supports the following:

* `index_document` - (Required) Storage returns this index document when requests are made to the root domain or any of the subfolders.

* `error_document` - (Optional) An absolute path to the document to return in case of a 4XX error.

The `CORS` object supports the following:

* `allowed_headers` (Optional) Specifies which headers are allowed.

* `allowed_methods` (Required) Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.

* `allowed_origins` (Required) Specifies which origins are allowed.

* `expose_headers` (Optional) Specifies expose header in the response.

* `max_age_seconds` (Optional) Specifies time in seconds that browser can cache the response for a preflight request.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The name of the bucket.

* `bucket_domain_name` - The bucket domain name.

* `website_endpoint` - The website endpoint, if the bucket is configured with a website. If not, this will be an empty string.

* `website_domain` - The domain of the website endpoint, if the bucket is configured with a website. If not, this will be an empty string.

## Import

Storage bucket can be imported using the `bucket`, e.g.

```
$ terraform import yandex_storage_bucket.bucket bucket-name
```

~> **Note:** Terraform will import this resource with `force_destroy` set to
`false` in state. If you've set it to `true` in config, run `terraform apply` to
update the value set in state. If you delete this resource before updating the
value, objects in the bucket will not be destroyed.
