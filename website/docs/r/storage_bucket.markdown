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

### Using object lifecycle

```hcl
resource "yandex_storage_bucket" "bucket" {
  bucket = "my-bucket"
  acl    = "private"

  lifecycle_rule {
    id      = "log"
    enabled = true

    prefix = "log/"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id      = "tmp"
    prefix  = "tmp/"
    enabled = true

    expiration {
      date = "2020-12-21"
    }
  }
}

resource "yandex_storage_bucket" "versioning_bucket" {
  bucket = "my-versioning-bucket"
  acl    = "private"

  versioning {
    enabled = true
  }

  lifecycle_rule {
    prefix  = "config/"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }
}
```

### Using SSE

```hcl
resource "yandex_kms_symmetric_key" "key-a" {
  name              = "example-symetric-key"
  description       = "description for key"
  default_algorithm = "AES_128"
  rotation_period   = "8760h" // equal to 1 year
}

resource "yandex_storage_bucket" "test" {
  bucket = "mybucket"

  server_side_encryption_configuration {
    rule {
  	  apply_server_side_encryption_by_default {
	    kms_master_key_id = yandex_kms_symmetric_key.key-a.id
	    sse_algorithm     = "aws:kms"
  	  }
    }
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

~> **Note:** To change ACL after creation, service account with `admin` role should be used, though this role is not necessary to create a bucket with any ACL.

* `grant` - (Optional) An [ACL policy grant](https://cloud.yandex.com/docs/storage/concepts/acl#permissions-types). Conflicts with `acl`.

~> **Note:** To manage `grant` argument, service account with `admin` role should be used.

* `force_destroy` - (Optional, Default: `false`) A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable.

* `website` - (Optional) A website object (documented below).

* `cors_rule` - (Optional) A rule of [Cross-Origin Resource Sharing](https://cloud.yandex.com/docs/storage/cors/) (documented below).

* `lifecycle_rule` - (Optional) A configuration of [object lifecycle management](https://cloud.yandex.com/docs/storage/concepts/lifecycles) (documented below).

The `website` object supports the following:

* `index_document` - (Required) Storage returns this index document when requests are made to the root domain or any of the subfolders.

* `error_document` - (Optional) An absolute path to the document to return in case of a 4XX error.

The `CORS` object supports the following:

* `allowed_headers` - (Optional) Specifies which headers are allowed.

* `allowed_methods` - (Required) Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.

* `allowed_origins` - (Required) Specifies which origins are allowed.

* `expose_headers` - (Optional) Specifies expose header in the response.

* `max_age_seconds` - (Optional) Specifies time in seconds that browser can cache the response for a preflight request.

* `server_side_encryption_configuration` - (Optional) A configuration of server-side encryption for the bucket (documented below)

The `lifecycle_rule` object supports the following:

* `id` - (Optional) Unique identifier for the rule. Must be less than or equal to 255 characters in length.

* `prefix` - (Optional) Object key prefix identifying one or more objects to which the rule applies.

* `enabled` - (Required) Specifies lifecycle rule status.

* `abort_incomplete_multipart_upload_days` - (Optional) Specifies the number of days after initiating a multipart upload when the multipart upload must be completed.

* `expiration` - (Optional) Specifies a period in the object's expire (documented below).

* `transition` - (Optional) Specifies a period in the object's transitions (documented below).

* `noncurrent_version_expiration` - (Optional) Specifies when noncurrent object versions expire (documented below).

* `noncurrent_version_transition` - (Optional) Specifies when noncurrent object versions transitions (documented below).

At least one of `abort_incomplete_multipart_upload_days`, `expiration`, `transition`, `noncurrent_version_expiration`, `noncurrent_version_transition` must be specified.

The `expiration` object supports the following

* `date` - (Optional) Specifies the date after which you want the corresponding action to take effect.

* `days` - (Optional) Specifies the number of days after object creation when the specific rule action takes effect.

* `expired_object_delete_marker` - (Optional) On a versioned bucket (versioning-enabled or versioning-suspended bucket), you can add this element in the lifecycle configuration to direct Object Storage to delete expired object delete markers.

The `transition` object supports the following

* `date` - (Optional) Specifies the date after which you want the corresponding action to take effect.

* `days` - (Optional) Specifies the number of days after object creation when the specific rule action takes effect.

* `storage_class` - (Required) Specifies the storage class to which you want the object to transition. Can only be `STANDARD_IA`.

The `noncurrent_version_expiration` object supports the following

* `days` - (Required) Specifies the number of days noncurrent object versions expire.

The `noncurrent_version_transition` object supports the following

* `days` - (Required) Specifies the number of days noncurrent object versions transition.

* `storage_class` - (Required) Specifies the storage class to which you want the noncurrent object versions to transition. Can only be `STANDARD_IA`.

The `server_side_encryption_configuration` object supports the following:

* `rule` - (Required) A single object for server-side encryption by default configuration. (documented below)

The `rule` object supports the following:

* `apply_server_side_encryption_by_default` - (Required) A single object for setting server-side encryption by default. (documented below)

The `apply_server_side_encryption_by_default` object supports the following:

* `sse_algorithm` - (Required) The server-side encryption algorithm to use. Single valid value is `aws:kms`

* `kms_master_key_id` - (Optional) The KMS master key ID used for the SSE-KMS encryption.

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
