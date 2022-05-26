---
layout: "yandex"
page_title: "Yandex: yandex_storage_bucket"
sidebar_current: "docs-yandex-storage-bucket"
description: |-
 Allows management of a Yandex.Cloud Storage Bucket.
---

# yandex\_storage\_bucket

Allows management of [Yandex.Cloud Storage Bucket](https://cloud.yandex.com/docs/storage/concepts/bucket).

~> **Note:** Your need to provide [static access key](https://cloud.yandex.com/docs/iam/concepts/authorization/access-key) (Access and Secret) to create storage client to work with Storage Service. To create them you need Service Account and proper permissions.

-> **Note:** For extended API usage, such as setting `max_size`, `folder_id`, `anonymous_access_flags`,
`default_storage_class` and `https` parameters for bucket, will be used default authorization method, i.e.
`IAM` / `OAuth` token from `provider` block will be used.
This might be a little bit confusing in cases when separate service account is used for managing buckets because
in this case buckets will be accessed by two different accounts that might have different permissions for buckets.

## Example Usage

### Simple Private Bucket

```hcl
locals {
  folder_id = "<folder-id>"
}

provider "yandex" {
  folder_id = local.folder_id
  zone      = "ru-central1-a"
}

// Create SA
resource "yandex_iam_service_account" "sa" {
  folder_id = local.folder_id
  name      = "tf-test-sa"
}

// Grant permissions
resource "yandex_resourcemanager_folder_iam_member" "sa-editor" {
  folder_id = local.folder_id
  role      = "storage.editor"
  member    = "serviceAccount:${yandex_iam_service_account.sa.id}"
}

// Create Static Access Keys
resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  service_account_id = yandex_iam_service_account.sa.id
  description        = "static access key for object storage"
}

// Use keys to create bucket
resource "yandex_storage_bucket" "test" {
  access_key = yandex_iam_service_account_static_access_key.sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-static-key.secret_key
  bucket = "tf-test-bucket"
}
```

### Static Website Hosting

```hcl
resource "yandex_storage_bucket" "test" {
  bucket = "storage-website-test.hashicorp.com"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
    routing_rules = <<EOF
[{
    "Condition": {
        "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
        "ReplaceKeyPrefixWith": "documents/"
    }
}]
EOF
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

### Using CORS

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "s3-website-test.hashicorp.com"
  acl    = "public-read"

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://s3-website-test.hashicorp.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

### Using versioning

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  versioning {
    enabled = true
  }
}
```

### Enable Logging

```hcl
resource "yandex_storage_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"
}

resource "yandex_storage_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  logging {
    target_bucket = yandex_storage_bucket.log_bucket.id
    target_prefix = "log/"
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
      storage_class = "COLD"
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
      storage_class = "COLD"
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

### Bucket Policy

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::my-policy-bucket/*",
        "arn:aws:s3:::my-policy-bucket"
      ]
    },
    {
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": [
        "arn:aws:s3:::my-policy-bucket/*",
        "arn:aws:s3:::my-policy-bucket"
      ]
    }
  ]
}
POLICY
}
```

### Bucket Max Size

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  max_size = 1048576
}
```

### Bucket Folder Id

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  folder_id = "<folder_id>"
}
```

### Bucket Anonymous Access Flags

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  anonymous_access_flags {
    read = true
    list = false
  }
}
```

### Bucket HTTPS Certificate

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  https {
    certificate_id = "<certificate_id_from_certificate_manager>"
  }
}
```

### Bucket Default Storage Class

```hcl
resource "yandex_storage_bucket" "b" {
  bucket = "my-policy-bucket"

  default_storage_class = "COLD"
}
```

### All settings example

```hcl
provider "yandex" {
  token = "<iam-token>"
  folder_id = "<folder-id>"
  storage_access_key = "<storage-access-key>"
  storage_secret_key = "<storage-secret-key>"
}

resource "yandex_storage_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"

  lifecycle_rule {
    id      = "cleanupoldlogs"
    enabled = true
    expiration {
      days = 365
    }
  }
}

resource "yandex_kms_symmetric_key" "key-a" {
  name              = "example-symetric-key"
  description       = "description for key"
  default_algorithm = "AES_128"
  rotation_period   = "8760h" // equal to 1 year
}

resource "yandex_storage_bucket" "all_settings" {
  bucket = "example-tf-settings-bucket"
  website {
    index_document = "index.html"
    error_document = "error.html"
  }

  lifecycle_rule {
    id = "test"
    enabled = true
    prefix = "prefix/"
    expiration {
      days = 30
    }
  }
  lifecycle_rule {
    id      = "log"
    enabled = true

    prefix = "log/"

    transition {
      days          = 30
      storage_class = "COLD"
    }

    expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id      = "everything180"
    prefix  = ""
    enabled = true

    expiration {
      days = 180
    }
  }
  lifecycle_rule {
    id      = "cleanupoldversions"
    prefix  = "config/"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "COLD"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }
  lifecycle_rule {
    id      = "abortmultiparts"
    prefix  = ""
    enabled = true
    abort_incomplete_multipart_upload_days = 7
  }

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT"]
    allowed_origins = ["https://storage-cloud.example.com"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }

  versioning {
    enabled = true
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = yandex_kms_symmetric_key.key-a.id
        sse_algorithm     = "aws:kms"
      }
    }
  }

  logging {
    target_bucket = yandex_storage_bucket.log_bucket.id
    target_prefix = "tf-logs/"
  }

  max_size = 1024

  folder_id = "<folder_id>"

  default_storage_class = "COLD"

  anonymous_access_flags {
    read = true
    list = true
  }

  https = {
    certificate_id = "<certificate_id>"
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

~> **Note:** To change ACL after creation, service account with `storage.admin` role should be used, though this role is not necessary to create a bucket with any ACL.

* `grant` - (Optional) An [ACL policy grant](https://cloud.yandex.com/docs/storage/concepts/acl#permissions-types). Conflicts with `acl`.

~> **Note:** To manage `grant` argument, service account with `storage.admin` role should be used.

* `force_destroy` - (Optional, Default: `false`) A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable.

* `website` - (Optional) A [website object](https://cloud.yandex.com/docs/storage/concepts/hosting) (documented below).

* `cors_rule` - (Optional) A rule of [Cross-Origin Resource Sharing](https://cloud.yandex.com/docs/storage/cors/) (documented below).

* `versioning` - (Optional) A state of [versioning](https://cloud.yandex.com/docs/storage/concepts/versioning) (documented below)

~> **Note:** To manage `versioning` argument, service account with `storage.admin` role should be used.

* `logging` - (Optional) A settings of [bucket logging](https://cloud.yandex.com/docs/storage/concepts/server-logs) (documented below).

* `lifecycle_rule` - (Optional) A configuration of [object lifecycle management](https://cloud.yandex.com/docs/storage/concepts/lifecycles) (documented below).

The `website` object supports the following:

* `index_document` - (Required, unless using `redirect_all_requests_to`) Storage returns this index document when requests are made to the root domain or any of the subfolders.

* `error_document` - (Optional) An absolute path to the document to return in case of a 4XX error.

* `redirect_all_requests_to` - (Optional) A hostname to redirect all website requests for this bucket to. Hostname can optionally be prefixed with a protocol (`http://` or `https://`) to use when redirecting requests. The default is the protocol that is used in the original request.

* `routing_rules` - (Optional) A json array containing [routing rules](https://cloud.yandex.com/docs/storage/s3/api-ref/hosting/upload#request-scheme) describing redirect behavior and when redirects are applied.

The `CORS` object supports the following:

* `allowed_headers` - (Optional) Specifies which headers are allowed.

* `allowed_methods` - (Required) Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.

* `allowed_origins` - (Required) Specifies which origins are allowed.

* `expose_headers` - (Optional) Specifies expose header in the response.

* `max_age_seconds` - (Optional) Specifies time in seconds that browser can cache the response for a preflight request.

* `server_side_encryption_configuration` - (Optional) A configuration of server-side encryption for the bucket (documented below)

The `versioning` object supports the following:

* `enabled` - (Optional) Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state. You can, however, suspend versioning on that bucket.

The `logging` object supports the following:

* `target_bucket` - (Required) The name of the bucket that will receive the log objects.

* `target_prefix` - (Optional) To specify a key prefix for log objects.

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

* `storage_class` - (Required) Specifies the storage class to which you want the object to transition. Can only be `COLD` or `STANDARD_IA`.

The `noncurrent_version_expiration` object supports the following

* `days` - (Required) Specifies the number of days noncurrent object versions expire.

The `noncurrent_version_transition` object supports the following

* `days` - (Required) Specifies the number of days noncurrent object versions transition.

* `storage_class` - (Required) Specifies the storage class to which you want the noncurrent object versions to transition. Can only be `COLD` or `STANDARD_IA`.

The `server_side_encryption_configuration` object supports the following:

* `rule` - (Required) A single object for server-side encryption by default configuration. (documented below)

The `rule` object supports the following:

* `apply_server_side_encryption_by_default` - (Required) A single object for setting server-side encryption by default. (documented below)

The `apply_server_side_encryption_by_default` object supports the following:

* `sse_algorithm` - (Required) The server-side encryption algorithm to use. Single valid value is `aws:kms`

* `kms_master_key_id` - (Optional) The KMS master key ID used for the SSE-KMS encryption.

The `policy` object should contain the only field with the text of the policy. See [policy documentation](https://cloud.yandex.com/docs/storage/concepts/policy) for more information on policy format.

Extended parameters of the bucket:

-> **NOTE:** for this parameters, authorization by `IAM-token` will be used.

* `folder_id` - (Optional) Allow to create bucket in different folder.

-> **NOTE:** it will try to create bucket using `IAM-token`, not using `access keys`.

* `max_size` - (Optional) The size of bucket, in bytes. See [size limiting](https://cloud.yandex.com/en-ru/docs/storage/operations/buckets/limit-max-volume) for more information.

* `default_storage_class` - (Optional) Storage class which is used for storing objects by default.
Available values are: "STANDARD", "COLD". Default is `"STANDARD"`.
See [storage class](https://cloud.yandex.com/en-ru/docs/storage/concepts/storage-class) for more inforamtion.

* `anonymous_access_flags` - (Optional) Provides various access to objects.
See [bucket availability](https://cloud.yandex.com/en-ru/docs/storage/operations/buckets/bucket-availability)
for more infomation.

* `https` - (Optional) Manages https certificates for bucket. See [https](https://cloud.yandex.com/en-ru/docs/storage/operations/hosting/certificate) for more infomation.

The `anonymous_access_flags` object supports the following properties:

* `read` - (Optional) Allows to read objects in bucket anonymously.

* `list` - (Optional) Allows to list object in bucket anonymously.

The `https` object supports the following properties:

* `certificate_id` — Id of the certificate in Certificate Manager, that will be used for bucket.

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
