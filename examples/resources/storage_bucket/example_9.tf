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

  lifecycle_rule {
    id      = "test_tag"
    enabled = true

    filter {
      tag {
        key   = "key1"
        value = "value1"
      }
    }

    expiration {
      date = "2020-12-21"
    }
  }

  lifecycle_rule {
    id      = "test_object_size_greater_than"
    enabled = true

    filter {
      object_size_greater_than = 1000
    }

    expiration {
      date = "2020-12-21"
    }
  }

  lifecycle_rule {
    id      = "object_size_less_than"
    enabled = true

    filter {
      object_size_less_than = 30000
    }

    expiration {
      date = "2020-12-21"
    }
  }

  lifecycle_rule {
    id      = "test_filter"
    enabled = true

    filter {
      and {
        object_size_greater_than = 1000
        object_size_less_than    = 30000
        prefix                   = "path2/"
        tags = {
          key1 = "value1"
          key2 = "value2"
        }
      }
    }

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
