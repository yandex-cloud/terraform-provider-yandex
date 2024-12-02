resource "yandex_serverless_container" "test-container" {
  name               = "some_name"
  description        = "any description"
  memory             = 256
  execution_timeout  = "15s"
  cores              = 1
  core_fraction      = 100
  service_account_id = "are1service2account3id"
  runtime {
    type = "task"
  }
  secrets {
    id                   = yandex_lockbox_secret.secret.id
    version_id           = yandex_lockbox_secret_version.secret_version.id
    key                  = "secret-key"
    environment_variable = "ENV_VARIABLE"
  }
  mounts {
    mount_point_path = "/mount/point"
    ephemeral_disk {
      size_gb = 5
    }
  }
  image {
    url = "cr.yandex/yc/test-image:v1"
  }
  log_options {
    log_group_id = "e2392vo6d1bne2aeq9fr"
    min_level    = "ERROR"
  }
  provision_policy {
    min_instances = 1
  }
}
