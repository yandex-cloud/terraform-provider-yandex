//
// Create a new Yandex Cloud Function
//
resource "yandex_function" "test-function" {
  name               = "some_name"
  description        = "any description"
  user_hash          = "any_user_defined_string"
  runtime            = "python37"
  entrypoint         = "main"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = "ajeih**********838kk"
  tags               = ["my_tag"]
  secrets {
    id                   = yandex_lockbox_secret.secret.id
    version_id           = yandex_lockbox_secret_version.secret_version.id
    key                  = "secret-key"
    environment_variable = "ENV_VARIABLE"
  }
  content {
    zip_filename = "function.zip"
  }
  mounts {
    name = "mnt"
    ephemeral_disk {
      size_gb = 32
    }
  }
  async_invocation {
    retries_count       = "3"
    service_account_id = "ajeih**********838kk"
    ymq_failure_target {
      service_account_id = "ajeqr**********qb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glr**********9hsfp:fail"
    }
    ymq_success_target {
      service_account_id = "ajeqr**********qb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glr**********9hsfp:success"
    }
  }
  log_options {
    log_group_id = "e2392**********eq9fr"
    min_level    = "ERROR"
  }
}
