resource "yandex_function" "test-function" {
  name               = "some_name"
  description        = "any description"
  user_hash          = "any_user_defined_string"
  runtime            = "python37"
  entrypoint         = "main"
  memory             = "128"
  execution_timeout  = "10"
  service_account_id = "are1service2account3id"
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
    services_account_id = "ajeihp9qsfg2l6f838kk"
    ymq_failure_target {
      service_account_id = "ajeqr0pjpbrkovcqb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glraqqa1i7tmh9hsfp:fail"
    }
    ymq_success_target {
      service_account_id = "ajeqr0pjpbrkovcqb76m"
      arn                = "yrn:yc:ymq:ru-central1:b1glraqqa1i7tmh9hsfp:success"
    }
  }
  log_options {
    log_group_id = "e2392vo6d1bne2aeq9fr"
    min_level    = "ERROR"
  }
}
