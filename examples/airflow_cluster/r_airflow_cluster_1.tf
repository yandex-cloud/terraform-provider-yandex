//
// Create a new Airflow Cluster.
//
resource "yandex_airflow_cluster" "my_airflow_cluster" {
  name               = "airflow-created-with-terraform"
  subnet_ids         = [yandex_vpc_subnet.a.id, yandex_vpc_subnet.b.id, yandex_vpc_subnet.d.id]
  service_account_id = yandex_iam_service_account.for-airflow.id
  admin_password     = "some-strong-password"

  code_sync = {
    s3 = {
      bucket = "bucket-for-airflow-dags"
    }
  }

  webserver = {
    count              = 1
    resource_preset_id = "c1-m4"
  }

  scheduler = {
    count              = 1
    resource_preset_id = "c1-m4"
  }

  worker = {
    min_count          = 1
    max_count          = 2
    resource_preset_id = "c1-m4"
  }

  airflow_config = {
    "api" = {
      "auth_backends" = "airflow.api.auth.backend.basic_auth,airflow.api.auth.backend.session"
    }
  }

  pip_packages = ["dbt"]

  lockbox_secrets_backend = {
    enabled = true
  }

  logging = {
    enabled   = true
    folder_id = var.folder_id
    min_level = "INFO"
  }
}
