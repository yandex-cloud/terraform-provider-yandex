resource "yandex_datatransfer_endpoint" "pg_source" {
  name = "pg-test-source"
  settings {
    postgres_source {
      connection {
        on_premise {
          hosts = [
            "example.org"
          ]
          port = 5432
        }
      }
      slot_gigabyte_lag_limit = 100
      database = "db1"
      user = "user1"
      password {
        raw = "123"
      }
    }
  }
}

resource "yandex_datatransfer_endpoint" "pg_target" {
  folder_id = "some_folder_id"
  name = "pg-test-target2"
  settings {
    postgres_target {
      connection {
        mdb_cluster_id = "some_cluster_id"
      }
      security_groups = [list of security group ids]
      database = "db2"
      user = "user2"
      password {
        raw = "321"
      }
    }
  }
}
