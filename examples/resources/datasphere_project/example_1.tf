resource "yandex_datasphere_project" "my-project" {
  name        = "example-datasphere-project"
  description = "Datasphere Project description"

  labels = {
    "foo" : "bar"
  }

  community_id = yandex_datasphere_community.my-community.id

  limits = {
    max_units_per_hour      = 10
    max_units_per_execution = 10
    balance                 = 10
  }

  settings = {
    service_account_id      = yandex_iam_service_account.my-account.id
    subnet_id               = yandex_vpc_subnet.my-subnet.id
    commit_mode             = "AUTO"
    data_proc_cluster_id    = "foo-data-proc-cluster-id"
    security_group_ids      = [yandex_vpc_security_group.my-security-group.id]
    ide                     = "JUPYTER_LAB"
    default_folder_id       = "foo-folder-id"
    stale_exec_timeout_mode = "ONE_HOUR"
  }
}
