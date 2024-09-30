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
      database = "db2"
      user = "user2"
      password {
        raw = "321"
      }
    }
  }
}

resource "yandex_datatransfer_transfer" "pgpg_transfer" {
  folder_id = "some_folder_id"
  name = "pgpg"
  source_id = yandex_datatransfer_endpoint.pg_source.id
  target_id = yandex_datatransfer_endpoint.pg_target.id
  type = "SNAPSHOT_AND_INCREMENT"
  runtime {
    yc_runtime {
      job_count = 1
      upload_shard_params {
        job_count = 4
        process_count = 1
      }
    }
  }
  transformation {
    transformers{
      one of transfomer
    }
    transformers{
      one of transfomers
    }
    ...
  }
}
